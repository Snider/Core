package main

import (
	"github.com/host-uk/core/cmd/core/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
