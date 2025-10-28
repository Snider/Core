package main

import (
	"github.com/Snider/Core/cmd/core/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
