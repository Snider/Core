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
	"github.com/Snider/Core"
	"github.com/Snider/Core/crypt"
)

app := core.New(
  core.WithService(crypt.Register),
  core.WithServiceLock(),
)
```

## Use
- Generate keys
- Encrypt/decrypt data
- Sign/verify messages

## Notes
- Uses [Proton Mail](https://pr.tn/ref/VZFX8H2VDCFG) OpenPGP fork.
