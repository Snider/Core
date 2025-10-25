package openpgp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// generateTestKeys creates a new PGP entity and saves the public and private keys to temporary files.
func generateTestKeys(t *testing.T, name, passphrase string) (string, string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "pgp-keys-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir for keys: %v", err)
	}

	entity, err := openpgp.NewEntity(name, "", name, nil)
	if err != nil {
		t.Fatalf("Failed to create new PGP entity: %v", err)
	}

	// Encrypt the private key with the passphrase
	if err := entity.PrivateKey.Encrypt([]byte(passphrase)); err != nil {
		t.Fatalf("Failed to encrypt private key: %v", err)
	}

	// --- Save Public Key ---
	pubKeyPath := filepath.Join(tempDir, name+".pub")
	pubKeyFile, err := os.Create(pubKeyPath)
	if err != nil {
		t.Fatalf("Failed to create public key file: %v", err)
	}
	w, err := armor.Encode(pubKeyFile, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("Failed to create armored writer for public key: %v", err)
	}
	if err := entity.Serialize(w); err != nil {
		t.Fatalf("Failed to serialize public key: %v", err)
	}
	w.Close()
	pubKeyFile.Close()

	// --- Save Private Key ---
	privKeyPath := filepath.Join(tempDir, name+".asc")
	privKeyFile, err := os.Create(privKeyPath)
	if err != nil {
		t.Fatalf("Failed to create private key file: %v", err)
	}
	w, err = armor.Encode(privKeyFile, openpgp.PrivateKeyType, nil)
	if err != nil {
		t.Fatalf("Failed to create armored writer for private key: %v", err)
	}
	if err := entity.SerializePrivate(w, nil); err != nil {
		t.Fatalf("Failed to serialize private key: %v", err)
	}
	w.Close()
	privKeyFile.Close()

	cleanup := func() { os.RemoveAll(tempDir) }
	return pubKeyPath, privKeyPath, cleanup
}

func TestEncryptDecryptPGP(t *testing.T) {
	recipientPub, recipientPriv, cleanup := generateTestKeys(t, "recipient", "recipient-pass")
	defer cleanup()

	originalMessage := "This is a secret message."

	// --- Test Encryption ---
	var encryptedBuf bytes.Buffer
	encryptedMessage, err := EncryptPGP(&encryptedBuf, recipientPub, originalMessage, nil, nil)
	if err != nil {
		t.Fatalf("EncryptPGP() failed: %v", err)
	}

	if !strings.Contains(encryptedMessage, "-----BEGIN PGP MESSAGE-----") {
		t.Errorf("Encrypted message does not appear to be PGP armored")
	}

	// --- Test Decryption ---
	decryptedMessage, err := DecryptPGP(recipientPriv, encryptedMessage, "recipient-pass", nil)
	if err != nil {
		t.Fatalf("DecryptPGP() failed: %v", err)
	}

	if decryptedMessage != originalMessage {
		t.Errorf("Decrypted message does not match original. got=%q, want=%q", decryptedMessage, originalMessage)
	}
}

func TestSignAndVerifyPGP(t *testing.T) {
	recipientPub, recipientPriv, rCleanup := generateTestKeys(t, "recipient", "recipient-pass")
	defer rCleanup()

	signerPub, signerPriv, sCleanup := generateTestKeys(t, "signer", "signer-pass")
	defer sCleanup()

	originalMessage := "This is a signed and verified message."

	// --- Encrypt and Sign ---
	var encryptedBuf bytes.Buffer
	signerPass := "signer-pass"
	encryptedMessage, err := EncryptPGP(&encryptedBuf, recipientPub, originalMessage, &signerPriv, &signerPass)
	if err != nil {
		t.Fatalf("EncryptPGP() with signing failed: %v", err)
	}

	// --- Decrypt and Verify ---
	decryptedMessage, err := DecryptPGP(recipientPriv, encryptedMessage, "recipient-pass", &signerPub)
	if err != nil {
		t.Fatalf("DecryptPGP() with verification failed: %v", err)
	}

	if decryptedMessage != originalMessage {
		t.Errorf("Decrypted message does not match original. got=%q, want=%q", decryptedMessage, originalMessage)
	}
}

func TestVerificationFailure(t *testing.T) {
	recipientPub, recipientPriv, rCleanup := generateTestKeys(t, "recipient", "recipient-pass")
	defer rCleanup()

	_, signerPriv, sCleanup := generateTestKeys(t, "signer", "signer-pass")
	defer sCleanup()

	// Generate a third, unexpected key to test verification failure
	unexpectedSignerPub, _, uCleanup := generateTestKeys(t, "unexpected", "unexpected-pass")
	defer uCleanup()

	originalMessage := "This message should fail verification."

	// --- Encrypt and Sign with the actual signer key ---
	var encryptedBuf bytes.Buffer
	signerPass := "signer-pass"
	encryptedMessage, err := EncryptPGP(&encryptedBuf, recipientPub, originalMessage, &signerPriv, &signerPass)
	if err != nil {
		t.Fatalf("EncryptPGP() with signing failed: %v", err)
	}

	// --- Attempt to Decrypt and Verify with the WRONG public key ---
	_, err = DecryptPGP(recipientPriv, encryptedMessage, "recipient-pass", &unexpectedSignerPub)
	if err == nil {
		t.Fatal("DecryptPGP() did not fail, but verification with an incorrect key was expected to fail.")
	}

	if !strings.Contains(err.Error(), "signature from unexpected key") {
		t.Errorf("Expected error to contain 'signature from unexpected key', but got: %v", err)
	}
}
