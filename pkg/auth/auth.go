// Package auth implements OpenPGP challenge-response authentication with
// support for both online (HTTP) and air-gapped (file-based) transport.
//
// Ported from dAppServer's mod-auth/lethean.service.ts.
//
// Authentication Flow (Online):
//
//  1. Client sends public key to server
//  2. Server generates a random nonce, encrypts it with client's public key
//  3. Client decrypts the nonce and signs it with their private key
//  4. Server verifies the signature, creates a session token
//
// Authentication Flow (Air-Gapped / Courier):
//
//	Same crypto but challenge/response are exchanged via files on a Medium.
//
// Storage Layout (via Medium):
//
//	users/
//	  {userID}.pub      PGP public key (armored)
//	  {userID}.key      PGP private key (armored, password-encrypted)
//	  {userID}.rev      Revocation certificate (placeholder)
//	  {userID}.json     User metadata (encrypted with user's public key)
//	  {userID}.lthn     LTHN password hash
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	coreerr "github.com/host-uk/core/pkg/framework/core"

	"github.com/host-uk/core/pkg/crypt/lthn"
	"github.com/host-uk/core/pkg/crypt/pgp"
	"github.com/host-uk/core/pkg/io"
)

// Default durations for challenge and session lifetimes.
const (
	DefaultChallengeTTL = 5 * time.Minute
	DefaultSessionTTL   = 24 * time.Hour
	nonceBytes          = 32
)

// protectedUsers lists usernames that cannot be deleted.
// The "server" user holds the server keypair; deleting it would
// permanently destroy all joining data and require a full rebuild.
var protectedUsers = map[string]bool{
	"server": true,
}

// User represents a registered user with PGP credentials.
type User struct {
	PublicKey    string    `json:"public_key"`
	KeyID        string    `json:"key_id"`
	Fingerprint  string    `json:"fingerprint"`
	PasswordHash string    `json:"password_hash"` // LTHN hash
	Created      time.Time `json:"created"`
	LastLogin    time.Time `json:"last_login"`
}

// Challenge is a PGP-encrypted nonce sent to a client during authentication.
type Challenge struct {
	Nonce     []byte    `json:"nonce"`
	Encrypted string    `json:"encrypted"`  // PGP-encrypted nonce (armored)
	ExpiresAt time.Time `json:"expires_at"`
}

// Session represents an authenticated session.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Option configures an Authenticator.
type Option func(*Authenticator)

// WithChallengeTTL sets the lifetime of a challenge before it expires.
func WithChallengeTTL(d time.Duration) Option {
	return func(a *Authenticator) {
		a.challengeTTL = d
	}
}

// WithSessionTTL sets the lifetime of a session before it expires.
func WithSessionTTL(d time.Duration) Option {
	return func(a *Authenticator) {
		a.sessionTTL = d
	}
}

// Authenticator manages PGP-based challenge-response authentication.
// All user data and keys are persisted through an io.Medium, which may
// be backed by disk, memory (MockMedium), or any other storage backend.
type Authenticator struct {
	medium       io.Medium
	sessions     map[string]*Session
	challenges   map[string]*Challenge // userID -> pending challenge
	mu           sync.RWMutex
	challengeTTL time.Duration
	sessionTTL   time.Duration
}

// New creates an Authenticator that persists user data via the given Medium.
func New(m io.Medium, opts ...Option) *Authenticator {
	a := &Authenticator{
		medium:       m,
		sessions:     make(map[string]*Session),
		challenges:   make(map[string]*Challenge),
		challengeTTL: DefaultChallengeTTL,
		sessionTTL:   DefaultSessionTTL,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// userPath returns the storage path for a user artifact.
func userPath(userID, ext string) string {
	return "users/" + userID + ext
}

// Register creates a new user account. It hashes the username with LTHN to
// produce a userID, generates a PGP keypair (protected by the given password),
// and persists the public key, private key, revocation placeholder, password
// hash, and encrypted metadata via the Medium.
func (a *Authenticator) Register(username, password string) (*User, error) {
	const op = "auth.Register"

	userID := lthn.Hash(username)

	// Check if user already exists
	if a.medium.IsFile(userPath(userID, ".pub")) {
		return nil, coreerr.E(op, "user already exists", nil)
	}

	// Ensure users directory exists
	if err := a.medium.EnsureDir("users"); err != nil {
		return nil, coreerr.E(op, "failed to create users directory", err)
	}

	// Generate PGP keypair
	kp, err := pgp.CreateKeyPair(userID, userID+"@auth.local", password)
	if err != nil {
		return nil, coreerr.E(op, "failed to create PGP keypair", err)
	}

	// Store public key
	if err := a.medium.Write(userPath(userID, ".pub"), kp.PublicKey); err != nil {
		return nil, coreerr.E(op, "failed to write public key", err)
	}

	// Store private key (already encrypted by PGP if password is non-empty)
	if err := a.medium.Write(userPath(userID, ".key"), kp.PrivateKey); err != nil {
		return nil, coreerr.E(op, "failed to write private key", err)
	}

	// Store revocation certificate placeholder
	if err := a.medium.Write(userPath(userID, ".rev"), "REVOCATION_PLACEHOLDER"); err != nil {
		return nil, coreerr.E(op, "failed to write revocation certificate", err)
	}

	// Store LTHN password hash
	passwordHash := lthn.Hash(password)
	if err := a.medium.Write(userPath(userID, ".lthn"), passwordHash); err != nil {
		return nil, coreerr.E(op, "failed to write password hash", err)
	}

	// Build user metadata
	now := time.Now()
	user := &User{
		PublicKey:    kp.PublicKey,
		KeyID:        userID,
		Fingerprint:  lthn.Hash(kp.PublicKey),
		PasswordHash: passwordHash,
		Created:      now,
		LastLogin:    time.Time{},
	}

	// Encrypt metadata with the user's public key and store
	metaJSON, err := json.Marshal(user)
	if err != nil {
		return nil, coreerr.E(op, "failed to marshal user metadata", err)
	}

	encMeta, err := pgp.Encrypt(metaJSON, kp.PublicKey)
	if err != nil {
		return nil, coreerr.E(op, "failed to encrypt user metadata", err)
	}

	if err := a.medium.Write(userPath(userID, ".json"), string(encMeta)); err != nil {
		return nil, coreerr.E(op, "failed to write user metadata", err)
	}

	return user, nil
}

// CreateChallenge generates a cryptographic challenge for the given user.
// A random nonce is created and encrypted with the user's PGP public key.
// The client must decrypt the nonce and sign it to prove key ownership.
func (a *Authenticator) CreateChallenge(userID string) (*Challenge, error) {
	const op = "auth.CreateChallenge"

	// Read user's public key
	pubKey, err := a.medium.Read(userPath(userID, ".pub"))
	if err != nil {
		return nil, coreerr.E(op, "user not found", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceBytes)
	if _, err := rand.Read(nonce); err != nil {
		return nil, coreerr.E(op, "failed to generate nonce", err)
	}

	// Encrypt nonce with user's public key
	encrypted, err := pgp.Encrypt(nonce, pubKey)
	if err != nil {
		return nil, coreerr.E(op, "failed to encrypt nonce", err)
	}

	challenge := &Challenge{
		Nonce:     nonce,
		Encrypted: string(encrypted),
		ExpiresAt: time.Now().Add(a.challengeTTL),
	}

	a.mu.Lock()
	a.challenges[userID] = challenge
	a.mu.Unlock()

	return challenge, nil
}

// ValidateResponse verifies a signed nonce from the client. The client must
// have decrypted the challenge nonce and signed it with their private key.
// On success, a new session is created and returned.
func (a *Authenticator) ValidateResponse(userID string, signedNonce []byte) (*Session, error) {
	const op = "auth.ValidateResponse"

	a.mu.Lock()
	challenge, exists := a.challenges[userID]
	if exists {
		delete(a.challenges, userID)
	}
	a.mu.Unlock()

	if !exists {
		return nil, coreerr.E(op, "no pending challenge for user", nil)
	}

	// Check challenge expiry
	if time.Now().After(challenge.ExpiresAt) {
		return nil, coreerr.E(op, "challenge expired", nil)
	}

	// Read user's public key
	pubKey, err := a.medium.Read(userPath(userID, ".pub"))
	if err != nil {
		return nil, coreerr.E(op, "user not found", err)
	}

	// Verify signature over the original nonce
	if err := pgp.Verify(challenge.Nonce, signedNonce, pubKey); err != nil {
		return nil, coreerr.E(op, "signature verification failed", err)
	}

	return a.createSession(userID)
}

// ValidateSession checks whether a token maps to a valid, non-expired session.
func (a *Authenticator) ValidateSession(token string) (*Session, error) {
	const op = "auth.ValidateSession"

	a.mu.RLock()
	session, exists := a.sessions[token]
	a.mu.RUnlock()

	if !exists {
		return nil, coreerr.E(op, "session not found", nil)
	}

	if time.Now().After(session.ExpiresAt) {
		a.mu.Lock()
		delete(a.sessions, token)
		a.mu.Unlock()
		return nil, coreerr.E(op, "session expired", nil)
	}

	return session, nil
}

// RefreshSession extends the expiry of an existing valid session.
func (a *Authenticator) RefreshSession(token string) (*Session, error) {
	const op = "auth.RefreshSession"

	a.mu.Lock()
	defer a.mu.Unlock()

	session, exists := a.sessions[token]
	if !exists {
		return nil, coreerr.E(op, "session not found", nil)
	}

	if time.Now().After(session.ExpiresAt) {
		delete(a.sessions, token)
		return nil, coreerr.E(op, "session expired", nil)
	}

	session.ExpiresAt = time.Now().Add(a.sessionTTL)
	return session, nil
}

// RevokeSession removes a session, invalidating the token immediately.
func (a *Authenticator) RevokeSession(token string) error {
	const op = "auth.RevokeSession"

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.sessions[token]; !exists {
		return coreerr.E(op, "session not found", nil)
	}

	delete(a.sessions, token)
	return nil
}

// DeleteUser removes a user and all associated keys from storage.
// The "server" user is protected and cannot be deleted (mirroring the
// original TypeScript implementation's safeguard).
func (a *Authenticator) DeleteUser(userID string) error {
	const op = "auth.DeleteUser"

	// Protect special users
	if protectedUsers[userID] {
		return coreerr.E(op, "cannot delete protected user", nil)
	}

	// Check user exists
	if !a.medium.IsFile(userPath(userID, ".pub")) {
		return coreerr.E(op, "user not found", nil)
	}

	// Remove all artifacts
	extensions := []string{".pub", ".key", ".rev", ".json", ".lthn"}
	for _, ext := range extensions {
		p := userPath(userID, ext)
		if a.medium.IsFile(p) {
			if err := a.medium.Delete(p); err != nil {
				return coreerr.E(op, "failed to delete "+ext, err)
			}
		}
	}

	// Revoke any active sessions for this user
	a.mu.Lock()
	for token, session := range a.sessions {
		if session.UserID == userID {
			delete(a.sessions, token)
		}
	}
	a.mu.Unlock()

	return nil
}

// Login performs password-based authentication as a convenience method.
// It verifies the password against the stored LTHN hash and, on success,
// creates a new session. This bypasses the PGP challenge-response flow.
func (a *Authenticator) Login(userID, password string) (*Session, error) {
	const op = "auth.Login"

	// Read stored password hash
	storedHash, err := a.medium.Read(userPath(userID, ".lthn"))
	if err != nil {
		return nil, coreerr.E(op, "user not found", err)
	}

	// Verify password
	if !lthn.Verify(password, storedHash) {
		return nil, coreerr.E(op, "invalid password", nil)
	}

	return a.createSession(userID)
}

// WriteChallengeFile writes an encrypted challenge to a file for air-gapped
// (courier) transport. The challenge is created and then its encrypted nonce
// is written to the specified path on the Medium.
func (a *Authenticator) WriteChallengeFile(userID, path string) error {
	const op = "auth.WriteChallengeFile"

	challenge, err := a.CreateChallenge(userID)
	if err != nil {
		return coreerr.E(op, "failed to create challenge", err)
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return coreerr.E(op, "failed to marshal challenge", err)
	}

	if err := a.medium.Write(path, string(data)); err != nil {
		return coreerr.E(op, "failed to write challenge file", err)
	}

	return nil
}

// ReadResponseFile reads a signed response from a file and validates it,
// completing the air-gapped authentication flow. The file must contain the
// raw PGP signature bytes (armored).
func (a *Authenticator) ReadResponseFile(userID, path string) (*Session, error) {
	const op = "auth.ReadResponseFile"

	content, err := a.medium.Read(path)
	if err != nil {
		return nil, coreerr.E(op, "failed to read response file", err)
	}

	session, err := a.ValidateResponse(userID, []byte(content))
	if err != nil {
		return nil, coreerr.E(op, "failed to validate response", err)
	}

	return session, nil
}

// createSession generates a cryptographically random session token and
// stores the session in the in-memory session map.
func (a *Authenticator) createSession(userID string) (*Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("auth: failed to generate session token: %w", err)
	}

	session := &Session{
		Token:     hex.EncodeToString(tokenBytes),
		UserID:    userID,
		ExpiresAt: time.Now().Add(a.sessionTTL),
	}

	a.mu.Lock()
	a.sessions[session.Token] = session
	a.mu.Unlock()

	return session, nil
}
