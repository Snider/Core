// Package lthn provides Lethean-specific cryptographic functions.
// It wraps the Enchantrix library's LTHN hash implementation.
package lthn

import (
	"github.com/Snider/Enchantrix/pkg/crypt"
)

var service *crypt.Service

func init() {
	service = crypt.NewService()
}

// Hash computes a Lethean-compatible hash of the input string.
// This is used for workspace identifiers and other obfuscation purposes.
func Hash(payload string) string {
	return service.Hash(crypt.LTHN, payload)
}
