// Package lthn implements the LTHN quasi-salted hash algorithm.
//
// LTHN produces deterministic, verifiable hashes without requiring separate salt
// storage. The salt is derived from the input itself through:
//  1. Reversing the input string
//  2. Applying "leet speak" style character substitutions
//
// The final hash is: SHA256(input || derived_salt)
//
// This is suitable for content identifiers, cache keys, and deduplication.
// NOT suitable for password hashing - use bcrypt, Argon2, or scrypt instead.
//
// Ported from Enchantrix (github.com/Snider/Enchantrix/pkg/crypt/std/lthn).
//
// Example:
//
//	hash := lthn.Hash("hello")
//	valid := lthn.Verify("hello", hash)  // true
package lthn

import (
	"crypto/sha256"
	"encoding/hex"
)

// keyMap defines the character substitutions for quasi-salt derivation.
// These are inspired by "leet speak" conventions for letter-number substitution.
// The mapping is bidirectional for most characters but NOT fully symmetric.
var keyMap = map[rune]rune{
	'o': '0', // letter O -> zero
	'l': '1', // letter L -> one
	'e': '3', // letter E -> three
	'a': '4', // letter A -> four
	's': 'z', // letter S -> Z
	't': '7', // letter T -> seven
	'0': 'o', // zero -> letter O
	'1': 'l', // one -> letter L
	'3': 'e', // three -> letter E
	'4': 'a', // four -> letter A
	'7': 't', // seven -> letter T
}

// SetKeyMap replaces the default character substitution map.
// Use this to customize the quasi-salt derivation for specific applications.
// Changes affect all subsequent Hash and Verify calls.
func SetKeyMap(newKeyMap map[rune]rune) {
	keyMap = newKeyMap
}

// GetKeyMap returns the current character substitution map.
func GetKeyMap() map[rune]rune {
	return keyMap
}

// Hash computes the LTHN hash of the input string.
//
// The algorithm:
//  1. Derive a quasi-salt by reversing the input and applying character substitutions
//  2. Concatenate: input + salt
//  3. Compute SHA-256 of the concatenated string
//  4. Return the hex-encoded digest (64 characters, lowercase)
//
// The same input always produces the same hash, enabling verification
// without storing a separate salt value.
func Hash(input string) string {
	salt := createSalt(input)
	hash := sha256.Sum256([]byte(input + salt))
	return hex.EncodeToString(hash[:])
}

// Verify checks if an input string produces the given hash.
// Returns true if Hash(input) equals the provided hash value.
func Verify(input string, hash string) bool {
	return Hash(input) == hash
}

// createSalt derives a quasi-salt by reversing the input and applying substitutions.
// For example: "hello" -> reversed "olleh" -> substituted "011eh"
func createSalt(input string) string {
	if input == "" {
		return ""
	}
	runes := []rune(input)
	salt := make([]rune, len(runes))
	for i := 0; i < len(runes); i++ {
		char := runes[len(runes)-1-i]
		if replacement, ok := keyMap[char]; ok {
			salt[i] = replacement
		} else {
			salt[i] = char
		}
	}
	return string(salt)
}
