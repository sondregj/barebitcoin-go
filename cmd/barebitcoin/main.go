package main

import (
	"os"

	"github.com/sondregj/barebitcoin-go/cmd/barebitcoin/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
