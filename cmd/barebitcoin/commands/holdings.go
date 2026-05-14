package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var (
	includeDeleted bool
)

func init() {
	holdingsCmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "Include deleted accounts")
}

var holdingsCmd = &cobra.Command{
	Use:   "holdings",
	Short: "Fetch bitcoin account balances",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runHoldingsCmd(cmd.Context(), client, includeDeleted)
	},
}

func runHoldingsCmd(
	ctx context.Context,
	client *barebitcoin.HTTPClient,
	includeDeleted bool,
) error {
	user, err := client.GetBitcoinAccounts(ctx, includeDeleted)
	if err != nil {
		return err
	}
	fmt.Println("accounts {")
	for _, account := range user.Accounts {
		fmt.Println("  account {")
		fmt.Println("    id", account.ID)
		fmt.Println("    name", account.Name)
		fmt.Println("    available btc", account.AvailableBTC)
		fmt.Println("    total btc", account.TotalBTC)
		fmt.Println("    total nok", account.TotalNOK)
		fmt.Println("  }")
	}
	fmt.Println("}")
	fmt.Println("total btc", user.TotalBTC)
	fmt.Println("total nok", user.TotalNOK)
	return nil
}
