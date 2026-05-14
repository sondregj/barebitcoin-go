package commands

import (
	"bytes"
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

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "ID\tNAME\tAVAILABLE BTC\tTOTAL BTC\tTOTAL NOK")
	for _, account := range user.Accounts {
		fmt.Fprintf(w, "%s\t%s\t%g\t%g\t%g\n",
			account.ID,
			account.Name,
			account.AvailableBTC,
			account.TotalBTC,
			account.TotalNOK,
		)
	}
	flushTable(w, &buf)

	fmt.Printf("\ntotal btc  %g\ntotal nok  %g\n", user.TotalBTC, user.TotalNOK)

	fiat, err := client.GetFiatAccount(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("fiat nok   %g\n", fiat.AvailableNOK)
	return nil
}
