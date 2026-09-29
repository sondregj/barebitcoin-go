package commands

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var balanceIncludeDeleted bool

func init() {
	balanceCmd.Flags().BoolVar(&balanceIncludeDeleted, "include-deleted", false, "Include deleted accounts")
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Fetch NOK and bitcoin account balances for tax purposes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runBalanceCmd(cmd.Context(), client, balanceIncludeDeleted)
	},
}

func runBalanceCmd(ctx context.Context, client *barebitcoin.HTTPClient, includeDeleted bool) error {
	resp, err := client.GetTaxBalance(ctx, includeDeleted)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "ID\tNAME\tBALANCE BTC\tDELETED")
	for _, account := range resp.BitcoinAccounts {
		deleted := ""
		if account.DeleteTime != nil {
			deleted = account.DeleteTime.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			account.ID,
			account.Name,
			account.BalanceBTC,
			deleted,
		)
	}
	flushTable(w, &buf)

	fmt.Printf("\nbalance nok  %s\n", resp.BalanceNOK)
	return nil
}
