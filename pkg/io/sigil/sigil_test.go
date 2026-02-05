package sigil

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ReverseSigil
// ---------------------------------------------------------------------------

func TestReverseSigil_Good(t *testing.T) {
	s := &ReverseSigil{}

	out, err := s.In([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, []byte("olleh"), out)

	// Symmetric: Out does the same thing.
	restored, err := s.Out(out)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), restored)
}

func TestReverseSigil_Bad(t *testing.T) {
	s := &ReverseSigil{}

	// Empty input returns empty.
	out, err := s.In([]byte{})
	require.NoError(t, err)
	assert.Equal(t, []byte{}, out)
}

func TestReverseSigil_Ugly(t *testing.T) {
	s := &ReverseSigil{}

	// Nil input returns nil.
	out, err := s.In(nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	out, err = s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// HexSigil
// ---------------------------------------------------------------------------

func TestHexSigil_Good(t *testing.T) {
	s := &HexSigil{}
	data := []byte("hello world")

	encoded, err := s.In(data)
	require.NoError(t, err)
	assert.Equal(t, []byte(hex.EncodeToString(data)), encoded)

	decoded, err := s.Out(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestHexSigil_Bad(t *testing.T) {
	s := &HexSigil{}

	// Invalid hex input.
	_, err := s.Out([]byte("zzzz"))
	assert.Error(t, err)

	// Empty input.
	out, err := s.In([]byte{})
	require.NoError(t, err)
	assert.Equal(t, []byte{}, out)
}

func TestHexSigil_Ugly(t *testing.T) {
	s := &HexSigil{}

	out, err := s.In(nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	out, err = s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// Base64Sigil
// ---------------------------------------------------------------------------

func TestBase64Sigil_Good(t *testing.T) {
	s := &Base64Sigil{}
	data := []byte("composable transforms")

	encoded, err := s.In(data)
	require.NoError(t, err)
	assert.Equal(t, []byte(base64.StdEncoding.EncodeToString(data)), encoded)

	decoded, err := s.Out(encoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestBase64Sigil_Bad(t *testing.T) {
	s := &Base64Sigil{}

	// Invalid base64 (wrong padding).
	_, err := s.Out([]byte("!!!"))
	assert.Error(t, err)

	// Empty input.
	out, err := s.In([]byte{})
	require.NoError(t, err)
	assert.Equal(t, []byte{}, out)
}

func TestBase64Sigil_Ugly(t *testing.T) {
	s := &Base64Sigil{}

	out, err := s.In(nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	out, err = s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// GzipSigil
// ---------------------------------------------------------------------------

func TestGzipSigil_Good(t *testing.T) {
	s := &GzipSigil{}
	data := []byte("the quick brown fox jumps over the lazy dog")

	compressed, err := s.In(data)
	require.NoError(t, err)
	assert.NotEqual(t, data, compressed)

	decompressed, err := s.Out(compressed)
	require.NoError(t, err)
	assert.Equal(t, data, decompressed)
}

func TestGzipSigil_Bad(t *testing.T) {
	s := &GzipSigil{}

	// Invalid gzip data.
	_, err := s.Out([]byte("not gzip"))
	assert.Error(t, err)

	// Empty input compresses to a valid gzip stream.
	compressed, err := s.In([]byte{})
	require.NoError(t, err)
	assert.NotEmpty(t, compressed) // gzip header is always present

	decompressed, err := s.Out(compressed)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, decompressed)
}

func TestGzipSigil_Ugly(t *testing.T) {
	s := &GzipSigil{}

	out, err := s.In(nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	out, err = s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// JSONSigil
// ---------------------------------------------------------------------------

func TestJSONSigil_Good(t *testing.T) {
	s := &JSONSigil{Indent: false}
	data := []byte(`{  "key" :   "value"  }`)

	compacted, err := s.In(data)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"key":"value"}`), compacted)

	// Out is passthrough.
	passthrough, err := s.Out(compacted)
	require.NoError(t, err)
	assert.Equal(t, compacted, passthrough)
}

func TestJSONSigil_Good_Indent(t *testing.T) {
	s := &JSONSigil{Indent: true}
	data := []byte(`{"key":"value"}`)

	indented, err := s.In(data)
	require.NoError(t, err)
	assert.Contains(t, string(indented), "\n")
	assert.Contains(t, string(indented), "  ")
}

func TestJSONSigil_Bad(t *testing.T) {
	s := &JSONSigil{Indent: false}

	// Invalid JSON.
	_, err := s.In([]byte("not json"))
	assert.Error(t, err)
}

func TestJSONSigil_Ugly(t *testing.T) {
	s := &JSONSigil{Indent: false}

	// json.Compact on nil/empty will produce an error (invalid JSON).
	_, err := s.In(nil)
	assert.Error(t, err)

	// Out with nil is passthrough.
	out, err := s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// HashSigil
// ---------------------------------------------------------------------------

func TestHashSigil_Good(t *testing.T) {
	data := []byte("hash me")

	tests := []struct {
		name     string
		sigilName string
		size     int
	}{
		{"md5", "md5", md5.Size},
		{"sha1", "sha1", sha1.Size},
		{"sha256", "sha256", sha256.Size},
		{"sha512", "sha512", sha512.Size},
		{"sha224", "sha224", sha256.Size224},
		{"sha384", "sha384", sha512.Size384},
		{"sha512-224", "sha512-224", 28},
		{"sha512-256", "sha512-256", 32},
		{"sha3-224", "sha3-224", 28},
		{"sha3-256", "sha3-256", 32},
		{"sha3-384", "sha3-384", 48},
		{"sha3-512", "sha3-512", 64},
		{"ripemd160", "ripemd160", 20},
		{"blake2s-256", "blake2s-256", 32},
		{"blake2b-256", "blake2b-256", 32},
		{"blake2b-384", "blake2b-384", 48},
		{"blake2b-512", "blake2b-512", 64},
		{"md4", "md4", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSigil(tt.sigilName)
			require.NoError(t, err)

			hashed, err := s.In(data)
			require.NoError(t, err)
			assert.Len(t, hashed, tt.size)

			// Out is passthrough.
			passthrough, err := s.Out(hashed)
			require.NoError(t, err)
			assert.Equal(t, hashed, passthrough)
		})
	}
}

func TestHashSigil_Bad(t *testing.T) {
	// Unsupported hash constant.
	s := &HashSigil{Hash: 0}
	_, err := s.In([]byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestHashSigil_Ugly(t *testing.T) {
	// Hashing empty data should still produce a valid digest.
	s, err := NewSigil("sha256")
	require.NoError(t, err)

	hashed, err := s.In([]byte{})
	require.NoError(t, err)
	assert.Len(t, hashed, sha256.Size)
}

// ---------------------------------------------------------------------------
// NewSigil factory
// ---------------------------------------------------------------------------

func TestNewSigil_Good(t *testing.T) {
	names := []string{
		"reverse", "hex", "base64", "gzip", "json", "json-indent",
		"md4", "md5", "sha1", "sha224", "sha256", "sha384", "sha512",
		"ripemd160",
		"sha3-224", "sha3-256", "sha3-384", "sha3-512",
		"sha512-224", "sha512-256",
		"blake2s-256", "blake2b-256", "blake2b-384", "blake2b-512",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			s, err := NewSigil(name)
			require.NoError(t, err)
			assert.NotNil(t, s)
		})
	}
}

func TestNewSigil_Bad(t *testing.T) {
	_, err := NewSigil("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown sigil name")
}

func TestNewSigil_Ugly(t *testing.T) {
	_, err := NewSigil("")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Transmute / Untransmute
// ---------------------------------------------------------------------------

func TestTransmute_Good(t *testing.T) {
	data := []byte("round trip")

	hexSigil, err := NewSigil("hex")
	require.NoError(t, err)
	base64Sigil, err := NewSigil("base64")
	require.NoError(t, err)

	chain := []Sigil{hexSigil, base64Sigil}

	encoded, err := Transmute(data, chain)
	require.NoError(t, err)
	assert.NotEqual(t, data, encoded)

	decoded, err := Untransmute(encoded, chain)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestTransmute_Good_MultiSigil(t *testing.T) {
	data := []byte("multi sigil pipeline test data")

	reverseSigil, err := NewSigil("reverse")
	require.NoError(t, err)
	hexSigil, err := NewSigil("hex")
	require.NoError(t, err)
	base64Sigil, err := NewSigil("base64")
	require.NoError(t, err)

	chain := []Sigil{reverseSigil, hexSigil, base64Sigil}

	encoded, err := Transmute(data, chain)
	require.NoError(t, err)

	decoded, err := Untransmute(encoded, chain)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestTransmute_Good_GzipRoundTrip(t *testing.T) {
	data := []byte("compress then encode then decode then decompress")

	gzipSigil, err := NewSigil("gzip")
	require.NoError(t, err)
	hexSigil, err := NewSigil("hex")
	require.NoError(t, err)

	chain := []Sigil{gzipSigil, hexSigil}

	encoded, err := Transmute(data, chain)
	require.NoError(t, err)

	decoded, err := Untransmute(encoded, chain)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
}

func TestTransmute_Bad(t *testing.T) {
	// Transmute with a sigil that will fail: hex decode on non-hex input.
	hexSigil := &HexSigil{}

	// Calling Out (decode) with invalid input via manual chain.
	_, err := Untransmute([]byte("not-hex!!"), []Sigil{hexSigil})
	assert.Error(t, err)
}

func TestTransmute_Ugly(t *testing.T) {
	// Empty sigil chain is a no-op.
	data := []byte("unchanged")

	result, err := Transmute(data, nil)
	require.NoError(t, err)
	assert.Equal(t, data, result)

	result, err = Untransmute(data, nil)
	require.NoError(t, err)
	assert.Equal(t, data, result)

	// Nil data through a chain.
	hexSigil, _ := NewSigil("hex")
	result, err = Transmute(nil, []Sigil{hexSigil})
	require.NoError(t, err)
	assert.Nil(t, result)
}
