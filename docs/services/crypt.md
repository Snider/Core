# Crypt Service

The Crypt service (`pkg/crypt`) provides cryptographic utilities including hashing, checksums, RSA encryption, and PGP operations.

## Features

- Multiple hash algorithms (SHA512, SHA256, SHA1, MD5)
- Checksum functions (Fletcher, Luhn)
- RSA key generation and encryption
- PGP encryption, signing, and verification
- Symmetric PGP encryption

## Basic Usage

```go
import "github.com/Snider/Core/pkg/crypt"

// Standalone usage
crypto, err := crypt.New()

// With Core framework
c, _ := core.New(
    core.WithService(crypt.Register),
)
crypto := core.MustServiceFor[*crypt.Service](c, "crypt")
```

## Hashing

```go
// Available algorithms: SHA512, SHA256, SHA1, MD5, LTHN
hash := crypto.Hash(crypt.SHA256, "hello world")

// Check if string is valid hash algorithm
isValid := crypto.IsHashAlgo("sha256")
```

## Checksums

```go
// Luhn validation (credit card numbers)
isValid := crypto.Luhn("4532015112830366")

// Fletcher checksums
f16 := crypto.Fletcher16("data")
f32 := crypto.Fletcher32("data")
f64 := crypto.Fletcher64("data")
```

## RSA Encryption

```go
// Generate key pair (2048 or 4096 bits recommended)
publicKey, privateKey, err := crypto.GenerateRSAKeyPair(2048)

// Encrypt with public key
ciphertext, err := crypto.EncryptRSA(publicKey, "secret message")

// Decrypt with private key
plaintext, err := crypto.DecryptRSA(privateKey, ciphertext)
```

## PGP Encryption

### Key Generation

```go
// Generate PGP key pair
publicKey, privateKey, err := crypto.GeneratePGPKeyPair(
    "User Name",
    "user@example.com",
    "Key comment",
)
```

### Asymmetric Encryption

```go
// Encrypt for recipient
ciphertext, err := crypto.EncryptPGPToString(recipientPublicKey, "secret data")

// Decrypt with private key
plaintext, err := crypto.DecryptPGP(privateKey, ciphertext)
```

### Symmetric Encryption

```go
var buf bytes.Buffer
err := crypto.SymmetricallyEncryptPGP(&buf, "data", "passphrase")
```

### Signing & Verification

```go
// Sign data
signature, err := crypto.SignPGP(privateKey, "data to sign")

// Verify signature
err := crypto.VerifyPGP(publicKey, "data to sign", signature)
if err != nil {
    // Signature invalid
}
```

## Hash Types

| Constant | Algorithm |
|----------|-----------|
| `crypt.SHA512` | SHA-512 |
| `crypt.SHA256` | SHA-256 |
| `crypt.SHA1` | SHA-1 |
| `crypt.MD5` | MD5 |
| `crypt.LTHN` | Custom LTHN hash |

## Frontend Usage (TypeScript)

```typescript
import {
    Hash,
    GenerateRSAKeyPair,
    EncryptRSA,
    DecryptRSA
} from '@bindings/crypt/service';

// Hash data
const hash = await Hash("SHA256", "hello world");

// RSA encryption
const { publicKey, privateKey } = await GenerateRSAKeyPair(2048);
const encrypted = await EncryptRSA(publicKey, "secret");
const decrypted = await DecryptRSA(privateKey, encrypted);
```
