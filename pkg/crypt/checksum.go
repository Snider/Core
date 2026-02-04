package crypt

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"os"

	core "github.com/host-uk/core/pkg/framework/core"
)

// SHA256File computes the SHA-256 checksum of a file and returns it as a hex string.
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", core.E("crypt.SHA256File", "failed to open file", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", core.E("crypt.SHA256File", "failed to read file", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA512File computes the SHA-512 checksum of a file and returns it as a hex string.
func SHA512File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", core.E("crypt.SHA512File", "failed to open file", err)
	}
	defer func() { _ = f.Close() }()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", core.E("crypt.SHA512File", "failed to read file", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA256Sum computes the SHA-256 checksum of data and returns it as a hex string.
func SHA256Sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SHA512Sum computes the SHA-512 checksum of data and returns it as a hex string.
func SHA512Sum(data []byte) string {
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}
