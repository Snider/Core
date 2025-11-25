---
title: crypt
---
# Service: `crypt`

The `crypt` service provides cryptographic utilities for the application, including hashing, checksums, and PGP encryption/decryption.

## Types

### `type HashType`

`HashType` defines the supported hashing algorithms.

```go
type HashType string
```

## Methods

### `func EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error)`

`EncryptPGP` encrypts data for a specific recipient.
- **writer**: Optional output writer.
- **recipientPath**: Path to the recipient's public key.
- **data**: The data to encrypt.
- **signerPath**: Optional path to a private key to sign the message.
- **signerPassphrase**: Optional passphrase for the signing key.

Returns the encrypted data as a string.

### `func DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error)`

`DecryptPGP` decrypts a PGP message.
- **recipientPath**: Path to the private key for decryption.
- **message**: The encrypted message (armor encoded).
- **passphrase**: Passphrase for the private key.
- **signerPath**: Optional path to the sender's public key to verify the signature.

Returns the decrypted string.

### `func Hash(lib HashType, payload string) string`

`Hash` computes a hash of the payload using the specified algorithm (e.g., MD5, SHA256).

### `func Fletcher16(payload string) uint16`

`Fletcher16` computes the Fletcher-16 checksum of the payload.

### `func Fletcher32(payload string) uint32`

`Fletcher32` computes the Fletcher-32 checksum of the payload.

### `func Fletcher64(payload string) uint64`

`Fletcher64` computes the Fletcher-64 checksum of the payload.

### `func Luhn(payload string) bool`

`Luhn` validates a number string using the Luhn algorithm (commonly used for credit card numbers).
