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
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(invoiceCmd)
	rootCmd.AddCommand(buyCmd)
	rootCmd.AddCommand(sellCmd)
	rootCmd.AddCommand(ordersCmd)
	rootCmd.AddCommand(cancelCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(receiveCmd)
	// TODO: user info
}

func Execute() error {
	return rootCmd.Execute()
}
