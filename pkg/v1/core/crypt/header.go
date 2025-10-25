package crypt

import (
	core "github.com/Snider/Core"
)

// HashType defines the supported hashing algorithms.
type HashType string

const (
	LTHN   HashType = "lthn"
	SHA512 HashType = "sha512"
	SHA256 HashType = "sha256"
	SHA1   HashType = "sha1"
	MD5    HashType = "md5"
)

// Service provides cryptographic functions.
// It is the main entry point for all cryptographic operations
// and is bound to the frontend.

type API struct {
	core *core.Core
}
