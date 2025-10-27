---
title: Core.Crypt
---

# Core.Crypt

Short: Keys, encrypt/decrypt, sign/verify.

## Overview
Simple wrappers around OpenPGP for common crypto tasks.

## Setup
```go
import (
  core "github.com/Snider/Core"
  crypt "github.com/Snider/Core/pkg/crypt"
)

// In your application's startup
cryptService, err := crypt.New()
if err != nil {
  // handle error
}
app := core.New(
  core.WithService(cryptService),
  core.WithServiceLock(),
)
```

## Use
- Generate keys
- Encrypt/decrypt data
- Sign/verify messages

## API
- `New() (*Service, error)`
- `Register(c *core.Core) (any, error)`
- `(s *Service) Hash(lib HashType, payload string) string`
- `(s *Service) Luhn(payload string) bool`
- `(s *Service) Fletcher16(payload string) uint16`
- `(s *Service) Fletcher32(payload string) uint32`
- `(s *Service) Fletcher64(payload string) uint64`
- `(s *Service) EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error)`
- `(s *Service) DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error)`

## Notes
- Uses [Proton Mail](https://pr.tn/ref/VZFX8H2VDCFG) OpenPGP fork.
