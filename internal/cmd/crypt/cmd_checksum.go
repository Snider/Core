package crypt

import (
	"fmt"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/crypt"
)

// Checksum command flags
var (
	checksumSHA512 bool
	checksumVerify string
)

func addChecksumCommand(parent *cli.Command) {
	checksumCmd := cli.NewCommand("checksum", "Compute file checksum", "", func(cmd *cli.Command, args []string) error {
		return runChecksum(args[0])
	})
	checksumCmd.Args = cli.ExactArgs(1)

	cli.BoolFlag(checksumCmd, &checksumSHA512, "sha512", "", false, "Use SHA-512 instead of SHA-256")
	cli.StringFlag(checksumCmd, &checksumVerify, "verify", "", "", "Verify file against this hash")

	parent.AddCommand(checksumCmd)
}

func runChecksum(path string) error {
	var hash string
	var err error

	if checksumSHA512 {
		hash, err = crypt.SHA512File(path)
	} else {
		hash, err = crypt.SHA256File(path)
	}

	if err != nil {
		return cli.Wrap(err, "failed to compute checksum")
	}

	if checksumVerify != "" {
		if hash == checksumVerify {
			cli.Success(fmt.Sprintf("Checksum matches: %s", filepath.Base(path)))
			return nil
		}
		cli.Error(fmt.Sprintf("Checksum mismatch: %s", filepath.Base(path)))
		cli.Dim(fmt.Sprintf("  expected: %s", checksumVerify))
		cli.Dim(fmt.Sprintf("  got:      %s", hash))
		return cli.Err("checksum verification failed")
	}

	algo := "SHA-256"
	if checksumSHA512 {
		algo = "SHA-512"
	}

	fmt.Printf("%s  %s  (%s)\n", hash, path, algo)
	return nil
}
