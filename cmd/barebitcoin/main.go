package main

import (
	"os"

	"github.com/sondregj/barebitcoin-go/cmd/barebitcoin/commands"
)

var (
	// These are injected automatically by goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := commands.Execute(version, commit, date); err != nil {
		os.Exit(1)
	}
}
