// Package pgp provides OpenPGP key generation, encryption, decryption,
// signing, and verification using the ProtonMail go-crypto library.
//
// Ported from Enchantrix (github.com/Snider/Enchantrix/pkg/crypt/std/pgp).
package pgp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// KeyPair holds armored PGP public and private keys.
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// CreateKeyPair generates a new PGP key pair for the given identity.
// If password is non-empty, the private key is encrypted with it.
// Returns a KeyPair with armored public and private keys.
func CreateKeyPair(name, email, password string) (*KeyPair, error) {
	entity, err := openpgp.NewEntity(name, "", email, nil)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to create entity: %w", err)
	}

	// Sign all the identities
	for _, id := range entity.Identities {
		_ = id.SelfSignature.SignUserId(id.UserId.Id, entity.PrimaryKey, entity.PrivateKey, nil)
	}

	// Encrypt private key with password if provided
	if password != "" {
		err = entity.PrivateKey.Encrypt([]byte(password))
		if err != nil {
			return nil, fmt.Errorf("pgp: failed to encrypt private key: %w", err)
		}
		for _, subkey := range entity.Subkeys {
			err = subkey.PrivateKey.Encrypt([]byte(password))
			if err != nil {
				return nil, fmt.Errorf("pgp: failed to encrypt subkey: %w", err)
			}
		}
	}

	// Serialize public key
	pubKeyBuf := new(bytes.Buffer)
	pubKeyWriter, err := armor.Encode(pubKeyBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to create armored public key writer: %w", err)
	}
	if err := entity.Serialize(pubKeyWriter); err != nil {
		pubKeyWriter.Close()
		return nil, fmt.Errorf("pgp: failed to serialize public key: %w", err)
	}
	pubKeyWriter.Close()

	// Serialize private key
	privKeyBuf := new(bytes.Buffer)
	privKeyWriter, err := armor.Encode(privKeyBuf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to create armored private key writer: %w", err)
	}
	if password != "" {
		// Manual serialization to avoid re-signing encrypted keys
		if err := serializeEncryptedEntity(privKeyWriter, entity); err != nil {
			privKeyWriter.Close()
			return nil, fmt.Errorf("pgp: failed to serialize private key: %w", err)
		}
	} else {
		if err := entity.SerializePrivate(privKeyWriter, nil); err != nil {
			privKeyWriter.Close()
			return nil, fmt.Errorf("pgp: failed to serialize private key: %w", err)
		}
	}
	privKeyWriter.Close()

	return &KeyPair{
		PublicKey:  pubKeyBuf.String(),
		PrivateKey: privKeyBuf.String(),
	}, nil
}

// serializeEncryptedEntity manually serializes an entity with encrypted private keys
// to avoid the panic from re-signing encrypted keys.
func serializeEncryptedEntity(w io.Writer, e *openpgp.Entity) error {
	if err := e.PrivateKey.Serialize(w); err != nil {
		return err
	}
	for _, ident := range e.Identities {
		if err := ident.UserId.Serialize(w); err != nil {
			return err
		}
		if err := ident.SelfSignature.Serialize(w); err != nil {
			return err
		}
	}
	for _, subkey := range e.Subkeys {
		if err := subkey.PrivateKey.Serialize(w); err != nil {
			return err
		}
		if err := subkey.Sig.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

// Encrypt encrypts data for the recipient identified by their armored public key.
// Returns the encrypted data as armored PGP output.
func Encrypt(data []byte, publicKeyArmor string) ([]byte, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(publicKeyArmor)))
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to read public key ring: %w", err)
	}

	buf := new(bytes.Buffer)
	armoredWriter, err := armor.Encode(buf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to create armor encoder: %w", err)
	}

	w, err := openpgp.Encrypt(armoredWriter, keyring, nil, nil, nil)
	if err != nil {
		armoredWriter.Close()
		return nil, fmt.Errorf("pgp: failed to create encryption writer: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		w.Close()
		armoredWriter.Close()
		return nil, fmt.Errorf("pgp: failed to write data: %w", err)
	}
	w.Close()
	armoredWriter.Close()

	return buf.Bytes(), nil
}

// Decrypt decrypts armored PGP data using the given armored private key.
// If the private key is encrypted, the password is used to decrypt it first.
func Decrypt(data []byte, privateKeyArmor, password string) ([]byte, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(privateKeyArmor)))
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to read private key ring: %w", err)
	}

	// Decrypt the private key if it is encrypted
	for _, entity := range keyring {
		if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
			if err := entity.PrivateKey.Decrypt([]byte(password)); err != nil {
				return nil, fmt.Errorf("pgp: failed to decrypt private key: %w", err)
			}
		}
		for _, subkey := range entity.Subkeys {
			if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
				_ = subkey.PrivateKey.Decrypt([]byte(password))
			}
		}
	}

	// Decode armored message
	block, err := armor.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to decode armored message: %w", err)
	}

	md, err := openpgp.ReadMessage(block.Body, keyring, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to read message: %w", err)
	}

	plaintext, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to read plaintext: %w", err)
	}

	return plaintext, nil
}

// Sign creates an armored detached signature for the given data using
// the armored private key. If the key is encrypted, the password is used
// to decrypt it first.
func Sign(data []byte, privateKeyArmor, password string) ([]byte, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(privateKeyArmor)))
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to read private key ring: %w", err)
	}

	signer := keyring[0]
	if signer.PrivateKey == nil {
		return nil, fmt.Errorf("pgp: private key not found in keyring")
	}

	if signer.PrivateKey.Encrypted {
		if err := signer.PrivateKey.Decrypt([]byte(password)); err != nil {
			return nil, fmt.Errorf("pgp: failed to decrypt private key: %w", err)
		}
	}

	buf := new(bytes.Buffer)
	config := &packet.Config{}
	err = openpgp.ArmoredDetachSign(buf, signer, bytes.NewReader(data), config)
	if err != nil {
		return nil, fmt.Errorf("pgp: failed to sign message: %w", err)
	}

	return buf.Bytes(), nil
}

// Verify verifies an armored detached signature against the given data
// and armored public key. Returns nil if the signature is valid.
func Verify(data, signature []byte, publicKeyArmor string) error {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(publicKeyArmor)))
	if err != nil {
		return fmt.Errorf("pgp: failed to read public key ring: %w", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(signature), nil)
	if err != nil {
		return fmt.Errorf("pgp: signature verification failed: %w", err)
	}

	return nil
}
