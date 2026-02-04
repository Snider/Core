package crypt

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
)

// Keygen command flags
var (
	keygenLength int
	keygenHex    bool
	keygenBase64 bool
)

func addKeygenCommand(parent *cli.Command) {
	keygenCmd := cli.NewCommand("keygen", "Generate a random cryptographic key", "", func(cmd *cli.Command, args []string) error {
		return runKeygen()
	})

	cli.IntFlag(keygenCmd, &keygenLength, "length", "l", 32, "Key length in bytes")
	cli.BoolFlag(keygenCmd, &keygenHex, "hex", "", false, "Output as hex string")
	cli.BoolFlag(keygenCmd, &keygenBase64, "base64", "", false, "Output as base64 string")

	parent.AddCommand(keygenCmd)
}

func runKeygen() error {
	if keygenHex && keygenBase64 {
		return cli.Err("--hex and --base64 are mutually exclusive")
	}
	if keygenLength <= 0 || keygenLength > 1024 {
		return cli.Err("key length must be between 1 and 1024 bytes")
	}

	key := make([]byte, keygenLength)
	if _, err := rand.Read(key); err != nil {
		return cli.Wrap(err, "failed to generate random key")
	}

	switch {
	case keygenHex:
		fmt.Println(hex.EncodeToString(key))
	case keygenBase64:
		fmt.Println(base64.StdEncoding.EncodeToString(key))
	default:
		// Default to hex output
		fmt.Println(hex.EncodeToString(key))
	}

	return nil
}
