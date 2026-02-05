package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/crypt/lthn"
	"github.com/host-uk/core/pkg/crypt/pgp"
	"github.com/host-uk/core/pkg/io"
)

// helper creates a fresh Authenticator backed by MockMedium.
func newTestAuth(opts ...Option) (*Authenticator, *io.MockMedium) {
	m := io.NewMockMedium()
	a := New(m, opts...)
	return a, m
}

// --- Register ---

func TestRegister_Good(t *testing.T) {
	a, m := newTestAuth()

	user, err := a.Register("alice", "hunter2")
	require.NoError(t, err)
	require.NotNil(t, user)

	userID := lthn.Hash("alice")

	// Verify public key is stored
	assert.True(t, m.IsFile(userPath(userID, ".pub")))
	assert.True(t, m.IsFile(userPath(userID, ".key")))
	assert.True(t, m.IsFile(userPath(userID, ".rev")))
	assert.True(t, m.IsFile(userPath(userID, ".json")))
	assert.True(t, m.IsFile(userPath(userID, ".lthn")))

	// Verify user fields
	assert.NotEmpty(t, user.PublicKey)
	assert.Equal(t, userID, user.KeyID)
	assert.NotEmpty(t, user.Fingerprint)
	assert.Equal(t, lthn.Hash("hunter2"), user.PasswordHash)
	assert.False(t, user.Created.IsZero())
}

func TestRegister_Bad(t *testing.T) {
	a, _ := newTestAuth()

	// Register first time succeeds
	_, err := a.Register("bob", "pass1")
	require.NoError(t, err)

	// Duplicate registration should fail
	_, err = a.Register("bob", "pass2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}

func TestRegister_Ugly(t *testing.T) {
	a, _ := newTestAuth()

	// Empty username/password should still work (PGP allows it)
	user, err := a.Register("", "")
	require.NoError(t, err)
	require.NotNil(t, user)
}

// --- CreateChallenge ---

func TestCreateChallenge_Good(t *testing.T) {
	a, _ := newTestAuth()

	user, err := a.Register("charlie", "pass")
	require.NoError(t, err)

	challenge, err := a.CreateChallenge(user.KeyID)
	require.NoError(t, err)
	require.NotNil(t, challenge)

	assert.Len(t, challenge.Nonce, nonceBytes)
	assert.NotEmpty(t, challenge.Encrypted)
	assert.True(t, challenge.ExpiresAt.After(time.Now()))
}

func TestCreateChallenge_Bad(t *testing.T) {
	a, _ := newTestAuth()

	// Challenge for non-existent user
	_, err := a.CreateChallenge("nonexistent-user-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestCreateChallenge_Ugly(t *testing.T) {
	a, _ := newTestAuth()

	// Empty userID
	_, err := a.CreateChallenge("")
	assert.Error(t, err)
}

// --- ValidateResponse (full challenge-response flow) ---

func TestValidateResponse_Good(t *testing.T) {
	a, m := newTestAuth()

	// Register user
	_, err := a.Register("dave", "password123")
	require.NoError(t, err)

	userID := lthn.Hash("dave")

	// Create challenge
	challenge, err := a.CreateChallenge(userID)
	require.NoError(t, err)

	// Client-side: decrypt nonce, then sign it
	privKey, err := m.Read(userPath(userID, ".key"))
	require.NoError(t, err)

	decryptedNonce, err := pgp.Decrypt([]byte(challenge.Encrypted), privKey, "password123")
	require.NoError(t, err)
	assert.Equal(t, challenge.Nonce, decryptedNonce)

	signedNonce, err := pgp.Sign(decryptedNonce, privKey, "password123")
	require.NoError(t, err)

	// Validate response
	session, err := a.ValidateResponse(userID, signedNonce)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEmpty(t, session.Token)
	assert.Equal(t, userID, session.UserID)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestValidateResponse_Bad(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("eve", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("eve")

	// No pending challenge
	_, err = a.ValidateResponse(userID, []byte("fake-signature"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pending challenge")
}

func TestValidateResponse_Ugly(t *testing.T) {
	a, m := newTestAuth(WithChallengeTTL(1 * time.Millisecond))

	_, err := a.Register("frank", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("frank")

	// Create challenge and let it expire
	challenge, err := a.CreateChallenge(userID)
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	// Sign with valid key but expired challenge
	privKey, err := m.Read(userPath(userID, ".key"))
	require.NoError(t, err)

	signedNonce, err := pgp.Sign(challenge.Nonce, privKey, "pass")
	require.NoError(t, err)

	_, err = a.ValidateResponse(userID, signedNonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "challenge expired")
}

// --- ValidateSession ---

func TestValidateSession_Good(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("grace", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("grace")

	session, err := a.Login(userID, "pass")
	require.NoError(t, err)

	validated, err := a.ValidateSession(session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.Token, validated.Token)
	assert.Equal(t, userID, validated.UserID)
}

func TestValidateSession_Bad(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.ValidateSession("nonexistent-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestValidateSession_Ugly(t *testing.T) {
	a, _ := newTestAuth(WithSessionTTL(1 * time.Millisecond))

	_, err := a.Register("heidi", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("heidi")

	session, err := a.Login(userID, "pass")
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	_, err = a.ValidateSession(session.Token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

// --- RefreshSession ---

func TestRefreshSession_Good(t *testing.T) {
	a, _ := newTestAuth(WithSessionTTL(1 * time.Hour))

	_, err := a.Register("ivan", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("ivan")

	session, err := a.Login(userID, "pass")
	require.NoError(t, err)

	originalExpiry := session.ExpiresAt

	// Small delay to ensure time moves forward
	time.Sleep(2 * time.Millisecond)

	refreshed, err := a.RefreshSession(session.Token)
	require.NoError(t, err)
	assert.True(t, refreshed.ExpiresAt.After(originalExpiry))
}

func TestRefreshSession_Bad(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.RefreshSession("nonexistent-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestRefreshSession_Ugly(t *testing.T) {
	a, _ := newTestAuth(WithSessionTTL(1 * time.Millisecond))

	_, err := a.Register("judy", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("judy")

	session, err := a.Login(userID, "pass")
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	_, err = a.RefreshSession(session.Token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

// --- RevokeSession ---

func TestRevokeSession_Good(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("karl", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("karl")

	session, err := a.Login(userID, "pass")
	require.NoError(t, err)

	err = a.RevokeSession(session.Token)
	require.NoError(t, err)

	// Token should no longer be valid
	_, err = a.ValidateSession(session.Token)
	assert.Error(t, err)
}

func TestRevokeSession_Bad(t *testing.T) {
	a, _ := newTestAuth()

	err := a.RevokeSession("nonexistent-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestRevokeSession_Ugly(t *testing.T) {
	a, _ := newTestAuth()

	// Revoke empty token
	err := a.RevokeSession("")
	assert.Error(t, err)
}

// --- DeleteUser ---

func TestDeleteUser_Good(t *testing.T) {
	a, m := newTestAuth()

	_, err := a.Register("larry", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("larry")

	// Also create a session that should be cleaned up
	_, err = a.Login(userID, "pass")
	require.NoError(t, err)

	err = a.DeleteUser(userID)
	require.NoError(t, err)

	// All files should be gone
	assert.False(t, m.IsFile(userPath(userID, ".pub")))
	assert.False(t, m.IsFile(userPath(userID, ".key")))
	assert.False(t, m.IsFile(userPath(userID, ".rev")))
	assert.False(t, m.IsFile(userPath(userID, ".json")))
	assert.False(t, m.IsFile(userPath(userID, ".lthn")))

	// Session should be gone
	a.mu.RLock()
	sessionCount := 0
	for _, s := range a.sessions {
		if s.UserID == userID {
			sessionCount++
		}
	}
	a.mu.RUnlock()
	assert.Equal(t, 0, sessionCount)
}

func TestDeleteUser_Bad(t *testing.T) {
	a, _ := newTestAuth()

	// Protected user "server" cannot be deleted
	err := a.DeleteUser("server")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete protected user")
}

func TestDeleteUser_Ugly(t *testing.T) {
	a, _ := newTestAuth()

	// Non-existent user
	err := a.DeleteUser("nonexistent-user-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// --- Login ---

func TestLogin_Good(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("mallory", "secret")
	require.NoError(t, err)
	userID := lthn.Hash("mallory")

	session, err := a.Login(userID, "secret")
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEmpty(t, session.Token)
	assert.Equal(t, userID, session.UserID)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestLogin_Bad(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("nancy", "correct-password")
	require.NoError(t, err)
	userID := lthn.Hash("nancy")

	// Wrong password
	_, err = a.Login(userID, "wrong-password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestLogin_Ugly(t *testing.T) {
	a, _ := newTestAuth()

	// Login for non-existent user
	_, err := a.Login("nonexistent-user-id", "pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// --- WriteChallengeFile / ReadResponseFile (Air-Gapped) ---

func TestAirGappedFlow_Good(t *testing.T) {
	a, m := newTestAuth()

	_, err := a.Register("oscar", "airgap-pass")
	require.NoError(t, err)
	userID := lthn.Hash("oscar")

	// Write challenge to file
	challengePath := "transfer/challenge.json"
	err = a.WriteChallengeFile(userID, challengePath)
	require.NoError(t, err)
	assert.True(t, m.IsFile(challengePath))

	// Read challenge file to get the encrypted nonce (simulating courier)
	challengeData, err := m.Read(challengePath)
	require.NoError(t, err)

	var challenge Challenge
	err = json.Unmarshal([]byte(challengeData), &challenge)
	require.NoError(t, err)

	// Client-side: decrypt nonce and sign it
	privKey, err := m.Read(userPath(userID, ".key"))
	require.NoError(t, err)

	decryptedNonce, err := pgp.Decrypt([]byte(challenge.Encrypted), privKey, "airgap-pass")
	require.NoError(t, err)

	signedNonce, err := pgp.Sign(decryptedNonce, privKey, "airgap-pass")
	require.NoError(t, err)

	// Write signed response to file
	responsePath := "transfer/response.sig"
	err = m.Write(responsePath, string(signedNonce))
	require.NoError(t, err)

	// Server reads response file
	session, err := a.ReadResponseFile(userID, responsePath)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEmpty(t, session.Token)
	assert.Equal(t, userID, session.UserID)
}

func TestWriteChallengeFile_Bad(t *testing.T) {
	a, _ := newTestAuth()

	// Challenge for non-existent user
	err := a.WriteChallengeFile("nonexistent-user", "challenge.json")
	assert.Error(t, err)
}

func TestReadResponseFile_Bad(t *testing.T) {
	a, _ := newTestAuth()

	// Response file does not exist
	_, err := a.ReadResponseFile("some-user", "nonexistent-file.sig")
	assert.Error(t, err)
}

func TestReadResponseFile_Ugly(t *testing.T) {
	a, m := newTestAuth()

	_, err := a.Register("peggy", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("peggy")

	// Create a challenge
	_, err = a.CreateChallenge(userID)
	require.NoError(t, err)

	// Write garbage to response file
	responsePath := "transfer/bad-response.sig"
	err = m.Write(responsePath, "not-a-valid-signature")
	require.NoError(t, err)

	_, err = a.ReadResponseFile(userID, responsePath)
	assert.Error(t, err)
}

// --- Options ---

func TestWithChallengeTTL_Good(t *testing.T) {
	ttl := 30 * time.Second
	a, _ := newTestAuth(WithChallengeTTL(ttl))
	assert.Equal(t, ttl, a.challengeTTL)
}

func TestWithSessionTTL_Good(t *testing.T) {
	ttl := 2 * time.Hour
	a, _ := newTestAuth(WithSessionTTL(ttl))
	assert.Equal(t, ttl, a.sessionTTL)
}

// --- Full Round-Trip (Online Flow) ---

func TestFullRoundTrip_Good(t *testing.T) {
	a, m := newTestAuth()

	// 1. Register
	user, err := a.Register("quinn", "roundtrip-pass")
	require.NoError(t, err)
	require.NotNil(t, user)

	userID := lthn.Hash("quinn")

	// 2. Create challenge
	challenge, err := a.CreateChallenge(userID)
	require.NoError(t, err)

	// 3. Client decrypts + signs
	privKey, err := m.Read(userPath(userID, ".key"))
	require.NoError(t, err)

	nonce, err := pgp.Decrypt([]byte(challenge.Encrypted), privKey, "roundtrip-pass")
	require.NoError(t, err)

	sig, err := pgp.Sign(nonce, privKey, "roundtrip-pass")
	require.NoError(t, err)

	// 4. Server validates, issues session
	session, err := a.ValidateResponse(userID, sig)
	require.NoError(t, err)
	require.NotNil(t, session)

	// 5. Validate session
	validated, err := a.ValidateSession(session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.Token, validated.Token)

	// 6. Refresh session
	refreshed, err := a.RefreshSession(session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.Token, refreshed.Token)

	// 7. Revoke session
	err = a.RevokeSession(session.Token)
	require.NoError(t, err)

	// 8. Session should be invalid now
	_, err = a.ValidateSession(session.Token)
	assert.Error(t, err)
}

// --- Concurrent Access ---

func TestConcurrentSessions_Good(t *testing.T) {
	a, _ := newTestAuth()

	_, err := a.Register("ruth", "pass")
	require.NoError(t, err)
	userID := lthn.Hash("ruth")

	// Create multiple sessions concurrently
	const n = 10
	sessions := make(chan *Session, n)
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		go func() {
			s, err := a.Login(userID, "pass")
			if err != nil {
				errs <- err
				return
			}
			sessions <- s
		}()
	}

	for i := 0; i < n; i++ {
		select {
		case s := <-sessions:
			require.NotNil(t, s)
			// Validate each session
			_, err := a.ValidateSession(s.Token)
			assert.NoError(t, err)
		case err := <-errs:
			t.Fatalf("concurrent login failed: %v", err)
		}
	}
}
