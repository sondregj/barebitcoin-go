package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "barebitcoin",
	Short: "Bare Bitcoin CLI",
}

func init() {
	// bitcoin := "₿"
	// bitcoinSatoshi := "₿̰"

	rootCmd.AddCommand(priceCmd)
	rootCmd.AddCommand(holdingsCmd)
	rootCmd.AddCommand(invoiceCmd)
	// TODO: user info
}

func Execute() error {
	return rootCmd.Execute()
}
