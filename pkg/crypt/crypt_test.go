package crypt

import (
	"bytes"
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Constructor Tests ---

func TestNew(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		service, err := New()
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("returns independent instances", func(t *testing.T) {
		service1, err1 := New()
		service2, err2 := New()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotSame(t, service1, service2)
	})
}

func TestRegister(t *testing.T) {
	t.Run("registers with core successfully", func(t *testing.T) {
		coreInstance, err := core.New()
		require.NoError(t, err)

		service, err := Register(coreInstance)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("returns Service type with Runtime", func(t *testing.T) {
		coreInstance, err := core.New()
		require.NoError(t, err)

		service, err := Register(coreInstance)
		require.NoError(t, err)

		cryptService, ok := service.(*Service)
		assert.True(t, ok)
		assert.NotNil(t, cryptService.Runtime)
	})
}

// --- Hash Tests ---

func TestHash(t *testing.T) {
	s, _ := New()

	t.Run("LTHN hash", func(t *testing.T) {
		hash := s.Hash(LTHN, "hello")
		assert.NotEmpty(t, hash)
		// LTHN hash should be consistent
		hash2 := s.Hash(LTHN, "hello")
		assert.Equal(t, hash, hash2)
	})

	t.Run("SHA512 hash", func(t *testing.T) {
		hash := s.Hash(SHA512, "hello")
		// Known SHA512 hash for "hello"
		expected := "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
		assert.Equal(t, expected, hash)
	})

	t.Run("SHA256 hash", func(t *testing.T) {
		hash := s.Hash(SHA256, "hello")
		// Known SHA256 hash for "hello"
		expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
		assert.Equal(t, expected, hash)
	})

	t.Run("SHA1 hash", func(t *testing.T) {
		hash := s.Hash(SHA1, "hello")
		// Known SHA1 hash for "hello"
		expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
		assert.Equal(t, expected, hash)
	})

	t.Run("MD5 hash", func(t *testing.T) {
		hash := s.Hash(MD5, "hello")
		// Known MD5 hash for "hello"
		expected := "5d41402abc4b2a76b9719d911017c592"
		assert.Equal(t, expected, hash)
	})

	t.Run("default falls back to SHA256", func(t *testing.T) {
		hash := s.Hash("unknown", "hello")
		sha256Hash := s.Hash(SHA256, "hello")
		assert.Equal(t, sha256Hash, hash)
	})

	t.Run("empty string hash", func(t *testing.T) {
		hash := s.Hash(SHA256, "")
		// Known SHA256 hash for empty string
		expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		assert.Equal(t, expected, hash)
	})

	t.Run("hash with special characters", func(t *testing.T) {
		hash := s.Hash(SHA256, "hello!@#$%^&*()")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 produces 64 hex chars
	})

	t.Run("hash with unicode", func(t *testing.T) {
		hash := s.Hash(SHA256, "你好世界")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})

	t.Run("hash consistency", func(t *testing.T) {
		payload := "test payload for consistency check"
		for _, hashType := range []HashType{LTHN, SHA512, SHA256, SHA1, MD5} {
			hash1 := s.Hash(hashType, payload)
			hash2 := s.Hash(hashType, payload)
			assert.Equal(t, hash1, hash2, "Hash should be consistent for %s", hashType)
		}
	})
}

// --- Luhn Tests ---

func TestLuhn(t *testing.T) {
	s, _ := New()

	t.Run("valid Luhn numbers", func(t *testing.T) {
		validNumbers := []string{
			"79927398713",
			"4532015112830366", // Visa test number
			"6011514433546201", // Discover test number
			"371449635398431",  // Amex test number
			"30569309025904",   // Diners Club test number
		}
		for _, num := range validNumbers {
			assert.True(t, s.Luhn(num), "Expected %s to be valid", num)
		}
	})

	t.Run("invalid Luhn numbers", func(t *testing.T) {
		invalidNumbers := []string{
			"79927398714",
			"1234567890",
			"1111111111",
			"1234567891",
		}
		for _, num := range invalidNumbers {
			assert.False(t, s.Luhn(num), "Expected %s to be invalid", num)
		}
	})

	t.Run("all zeros is valid", func(t *testing.T) {
		// All zeros: each digit contributes 0, sum=0, 0%10==0
		assert.True(t, s.Luhn("0000000000"))
	})

	t.Run("handles spaces", func(t *testing.T) {
		// Same number with and without spaces should give same result
		assert.True(t, s.Luhn("7992 7398 713"))
		assert.True(t, s.Luhn("4532 0151 1283 0366"))
	})

	t.Run("non-digit characters return false", func(t *testing.T) {
		assert.False(t, s.Luhn("1234abcd5678"))
		assert.False(t, s.Luhn("12-34-56-78"))
		assert.False(t, s.Luhn("1234.5678"))
	})

	t.Run("empty string", func(t *testing.T) {
		// Empty string: sum=0, 0%10==0, so it returns true
		assert.True(t, s.Luhn(""))
	})

	t.Run("single digit", func(t *testing.T) {
		assert.True(t, s.Luhn("0"))
		assert.False(t, s.Luhn("1"))
	})
}

// --- Fletcher Checksum Tests ---

func TestFletcher16(t *testing.T) {
	s, _ := New()

	t.Run("basic checksum", func(t *testing.T) {
		checksum := s.Fletcher16("hello")
		assert.NotZero(t, checksum)
	})

	t.Run("empty string", func(t *testing.T) {
		checksum := s.Fletcher16("")
		assert.Equal(t, uint16(0), checksum)
	})

	t.Run("consistency", func(t *testing.T) {
		checksum1 := s.Fletcher16("test data")
		checksum2 := s.Fletcher16("test data")
		assert.Equal(t, checksum1, checksum2)
	})

	t.Run("different inputs produce different checksums", func(t *testing.T) {
		checksum1 := s.Fletcher16("hello")
		checksum2 := s.Fletcher16("world")
		assert.NotEqual(t, checksum1, checksum2)
	})

	t.Run("known value", func(t *testing.T) {
		// "abcde" has a known Fletcher-16 checksum
		checksum := s.Fletcher16("abcde")
		assert.NotZero(t, checksum)
	})
}

func TestFletcher32(t *testing.T) {
	s, _ := New()

	t.Run("basic checksum", func(t *testing.T) {
		checksum := s.Fletcher32("hello")
		assert.NotZero(t, checksum)
	})

	t.Run("empty string", func(t *testing.T) {
		checksum := s.Fletcher32("")
		assert.Equal(t, uint32(0), checksum)
	})

	t.Run("consistency", func(t *testing.T) {
		checksum1 := s.Fletcher32("test data")
		checksum2 := s.Fletcher32("test data")
		assert.Equal(t, checksum1, checksum2)
	})

	t.Run("different inputs produce different checksums", func(t *testing.T) {
		checksum1 := s.Fletcher32("hello")
		checksum2 := s.Fletcher32("world")
		assert.NotEqual(t, checksum1, checksum2)
	})

	t.Run("handles odd-length input", func(t *testing.T) {
		// Odd length input should be padded
		checksum := s.Fletcher32("abc")
		assert.NotZero(t, checksum)
	})

	t.Run("handles even-length input", func(t *testing.T) {
		checksum := s.Fletcher32("abcd")
		assert.NotZero(t, checksum)
	})
}

func TestFletcher64(t *testing.T) {
	s, _ := New()

	t.Run("basic checksum", func(t *testing.T) {
		checksum := s.Fletcher64("hello")
		assert.NotZero(t, checksum)
	})

	t.Run("empty string", func(t *testing.T) {
		checksum := s.Fletcher64("")
		assert.Equal(t, uint64(0), checksum)
	})

	t.Run("consistency", func(t *testing.T) {
		checksum1 := s.Fletcher64("test data")
		checksum2 := s.Fletcher64("test data")
		assert.Equal(t, checksum1, checksum2)
	})

	t.Run("different inputs produce different checksums", func(t *testing.T) {
		checksum1 := s.Fletcher64("hello")
		checksum2 := s.Fletcher64("world")
		assert.NotEqual(t, checksum1, checksum2)
	})

	t.Run("handles various input lengths", func(t *testing.T) {
		// Test padding for different lengths
		for i := 1; i <= 8; i++ {
			input := string(make([]byte, i))
			checksum := s.Fletcher64(input)
			// Just verify it doesn't panic
			_ = checksum
		}
	})

	t.Run("long input", func(t *testing.T) {
		// Use actual text content, not null bytes
		longInput := ""
		for i := 0; i < 100; i++ {
			longInput += "test data "
		}
		checksum := s.Fletcher64(longInput)
		assert.NotZero(t, checksum)
	})
}

// --- HashType Constants Tests ---

func TestHashTypeConstants(t *testing.T) {
	t.Run("constants have expected values", func(t *testing.T) {
		assert.Equal(t, HashType("lthn"), LTHN)
		assert.Equal(t, HashType("sha512"), SHA512)
		assert.Equal(t, HashType("sha256"), SHA256)
		assert.Equal(t, HashType("sha1"), SHA1)
		assert.Equal(t, HashType("md5"), MD5)
	})
}

// --- PGP Tests (basic, detailed tests in openpgp package) ---

func TestEncryptPGP(t *testing.T) {
	t.Run("requires valid key paths", func(t *testing.T) {
		s, _ := New()
		var buf bytes.Buffer

		// Should fail with invalid path
		_, err := s.EncryptPGP(&buf, "/nonexistent/path", "test data", nil, nil)
		assert.Error(t, err)
	})
}

func TestDecryptPGP(t *testing.T) {
	t.Run("requires valid key paths", func(t *testing.T) {
		s, _ := New()

		// Should fail with invalid path
		_, err := s.DecryptPGP("/nonexistent/path", "encrypted data", "passphrase", nil)
		assert.Error(t, err)
	})
}

// --- HandleIPCEvents Tests ---

func TestHandleIPCEvents(t *testing.T) {
	t.Run("handles ActionServiceStartup", func(t *testing.T) {
		coreInstance, err := core.New()
		require.NoError(t, err)

		serviceAny, err := Register(coreInstance)
		require.NoError(t, err)

		s := serviceAny.(*Service)
		err = s.HandleIPCEvents(coreInstance, core.ActionServiceStartup{})
		assert.NoError(t, err)
	})

	t.Run("handles unknown message type", func(t *testing.T) {
		coreInstance, err := core.New()
		require.NoError(t, err)

		serviceAny, err := Register(coreInstance)
		require.NoError(t, err)

		s := serviceAny.(*Service)
		// Pass an arbitrary type as unknown message
		err = s.HandleIPCEvents(coreInstance, "unknown message")
		assert.NoError(t, err)
	})
}
