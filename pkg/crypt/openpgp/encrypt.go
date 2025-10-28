package openpgp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// readRecipientEntity reads an armored PGP public key from the given path.
func readRecipientEntity(path string) (*openpgp.Entity, error) {
	recipientFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to open recipient public key file at %s: %w", path, err)
	}
	defer func(recipientFile *os.File) {
		if err := recipientFile.Close(); err != nil {
			fmt.Printf("openpgp: warning - failed to close recipient key file: %v\n", err)
		}
	}(recipientFile)

	block, err := armor.Decode(recipientFile)
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to decode armored key from %s: %w", path, err)
	}

	if block.Type != openpgp.PublicKeyType {
		return nil, fmt.Errorf("openpgp: invalid key type in %s: expected public key, got %s", path, block.Type)
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to read entity from public key: %w", err)
	}
	return entity, nil
}

// readSignerEntity reads and decrypts an armored PGP private key.
func readSignerEntity(path, passphrase string) (*openpgp.Entity, error) {
	signerFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to open signer private key file at %s: %w", path, err)
	}
	defer func(signerFile *os.File) {
		if err := signerFile.Close(); err != nil {
			fmt.Printf("openpgp: warning - failed to close signer key file: %v\n", err)
		}
	}(signerFile)

	block, err := armor.Decode(signerFile)
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to decode armored key from %s: %w", path, err)
	}

	if block.Type != openpgp.PrivateKeyType {
		return nil, fmt.Errorf("openpgp: invalid key type in %s: expected private key, got %s", path, block.Type)
	}

	entity, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	if err != nil {
		return nil, fmt.Errorf("openpgp: failed to read entity from private key: %w", err)
	}

	// Only decrypt if the key is actually encrypted.
	if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
		if err := entity.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
			return nil, fmt.Errorf("openpgp: failed to decrypt private key, check your passphrase: %w", err)
		}
	}

	return entity, nil
}

// EncryptPGP encrypts a string using PGP, writing the armored result to the writer.
func EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error) {
	// 1. Read the recipient's public key
	recipientEntity, err := readRecipientEntity(recipientPath)
	if err != nil {
		return "", err
	}

	// 2. Set up the list of recipients
	to := []*openpgp.Entity{recipientEntity}

	// 3. Handle optional signing
	var signer *openpgp.Entity
	if signerPath != nil {
		signer, err = readSignerEntity(*signerPath, *signerPassphrase)
		if err != nil {
			return "", fmt.Errorf("openpgp: failed to prepare signer: %w", err)
		}
	}

	// 4. Create an armored writer and encrypt the message
	armoredWriter, err := armor.Encode(writer, "PGP MESSAGE", nil)
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to create armored writer: %w", err)
	}

	plaintext, err := openpgp.Encrypt(armoredWriter, to, signer, nil, nil)
	if err != nil {
		_ = armoredWriter.Close() // Attempt to close, but prioritize the encryption error.
		return "", fmt.Errorf("openpgp: failed to begin encryption: %w", err)
	}

	_, err = plaintext.Write([]byte(data))
	if err != nil {
		_ = plaintext.Close()
		_ = armoredWriter.Close()
		return "", fmt.Errorf("openpgp: failed to write data to encryption stream: %w", err)
	}

	// 5. Explicitly close the writers to finalize the message before reading.
	if err := plaintext.Close(); err != nil {
		return "", fmt.Errorf("openpgp: failed to finalize plaintext writer: %w", err)
	}
	if err := armoredWriter.Close(); err != nil {
		return "", fmt.Errorf("openpgp: failed to finalize armored writer: %w", err)
	}

	// To get the string output, we need to have written to a buffer
	if buf, ok := writer.(*bytes.Buffer); ok {
		return buf.String(), nil
	}

	return "", nil // No string to return if not writing to a buffer
}

// DecryptPGP decrypts an armored PGP message.
func DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error) {
	// 1. Read the recipient's private key
	recipientFile, err := os.Open(recipientPath)
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to open recipient private key file at %s: %w", recipientPath, err)
	}
	defer func(recipientFile *os.File) {
		if err := recipientFile.Close(); err != nil {
			fmt.Printf("openpgp: warning - failed to close recipient key file: %v\n", err)
		}
	}(recipientFile)

	entityList, err := openpgp.ReadArmoredKeyRing(recipientFile)
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to read armored key ring from %s: %w", recipientPath, err)
	}
	if len(entityList) == 0 {
		return "", fmt.Errorf("openpgp: no keys found in recipient key file %s", recipientPath)
	}

	// 2. Decrypt the private key
	entity := entityList[0]
	if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
		if err := entity.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
			return "", fmt.Errorf("openpgp: failed to decrypt private key, check your passphrase: %w", err)
		}
	}

	// 3. Decode the armored message
	block, err := armor.Decode(strings.NewReader(message))
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to decode armored message: %w", err)
	}
	if block.Type != "PGP MESSAGE" {
		return "", fmt.Errorf("openpgp: invalid message type: got %s, want PGP MESSAGE", block.Type)
	}

	// 4. Decrypt the message body
	md, err := openpgp.ReadMessage(block.Body, entityList, nil, nil)
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to read PGP message: %w", err)
	}

	plaintext, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", fmt.Errorf("openpgp: failed to read plaintext message body: %w", err)
	}

	// 5. Handle optional signature verification
	if signerPath != nil {
		// First, ensure a signature actually exists when one is expected.
		if md.SignedByKeyId == 0 {
			return "", fmt.Errorf("openpgp: signature verification failed: message is not signed")
		}

		signer, err := readRecipientEntity(*signerPath) // Signer is verified with their public key
		if err != nil {
			return "", fmt.Errorf("openpgp: failed to read signer public key: %w", err)
		}
		if md.SignatureError != nil {
			return "", fmt.Errorf("openpgp: signature verification failed: %w", md.SignatureError)
		}
		if md.SignedByKeyId != signer.PrimaryKey.KeyId {
			return "", fmt.Errorf("openpgp: signature from unexpected key id: got %d, want %d", md.SignedByKeyId, signer.PrimaryKey.KeyId)
		}
	}

	return string(plaintext), nil
}
