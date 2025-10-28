// package crypt provides the public API for the crypt service.
package crypt

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/crypt"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Service service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// Options is the public type for the Options service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Options = impl.Options

// HashType is the public type for the HashType service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type HashType = impl.HashType

// LTHN is a public constant that points to the real constant in the implementation package.
const LTHN = impl.LTHN

// SHA512 is a public constant that points to the real constant in the implementation package.
const SHA512 = impl.SHA512

// SHA256 is a public constant that points to the real constant in the implementation package.
const SHA256 = impl.SHA256

// SHA1 is a public constant that points to the real constant in the implementation package.
const SHA1 = impl.SHA1

// MD5 is a public constant that points to the real constant in the implementation package.
const MD5 = impl.MD5

// New is a public function that points to the real function in the implementation package.
var New = impl.New

// Register is a public function that points to the real function in the implementation package.
var Register = impl.Register

// Crypt is the public interface for the crypt service.
type Crypt = core.Crypt
