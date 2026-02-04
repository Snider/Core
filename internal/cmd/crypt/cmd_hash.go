package crypt

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/crypt"
	"golang.org/x/crypto/bcrypt"
)

// Hash command flags
var (
	hashBcrypt bool
	hashVerify string
)

func addHashCommand(parent *cli.Command) {
	hashCmd := cli.NewCommand("hash", "Hash a password with Argon2id or bcrypt", "", func(cmd *cli.Command, args []string) error {
		return runHash(args[0])
	})
	hashCmd.Args = cli.ExactArgs(1)

	cli.BoolFlag(hashCmd, &hashBcrypt, "bcrypt", "b", false, "Use bcrypt instead of Argon2id")
	cli.StringFlag(hashCmd, &hashVerify, "verify", "", "", "Verify input against this hash")

	parent.AddCommand(hashCmd)
}

func runHash(input string) error {
	// Verify mode
	if hashVerify != "" {
		return runHashVerify(input, hashVerify)
	}

	// Hash mode
	if hashBcrypt {
		hash, err := crypt.HashBcrypt(input, bcrypt.DefaultCost)
		if err != nil {
			return cli.Wrap(err, "failed to hash password")
		}
		fmt.Println(hash)
		return nil
	}

	hash, err := crypt.HashPassword(input)
	if err != nil {
		return cli.Wrap(err, "failed to hash password")
	}
	fmt.Println(hash)
	return nil
}

func runHashVerify(input, hash string) error {
	var match bool
	var err error

	if hashBcrypt {
		match, err = crypt.VerifyBcrypt(input, hash)
	} else {
		match, err = crypt.VerifyPassword(input, hash)
	}

	if err != nil {
		return cli.Wrap(err, "failed to verify hash")
	}

	if match {
		cli.Success("Password matches hash")
		return nil
	}

	cli.Error("Password does not match hash")
	return cli.Err("hash verification failed")
}
