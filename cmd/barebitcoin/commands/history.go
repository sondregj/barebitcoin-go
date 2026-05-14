package commands

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var historyLimit int

func init() {
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Number of entries to show")
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Fetch transaction history",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runHistoryCmd(cmd.Context(), client, historyLimit)
	},
}

func runHistoryCmd(ctx context.Context, client *barebitcoin.HTTPClient, limit int) error {
	resp, err := client.GetTaxTransactions(ctx)
	if err != nil {
		return err
	}

	entries := resp.Transactions
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "ID\tTYPE\tCREATED\tIN\tOUT\tFEE")
	for _, entry := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s %s\t%s %s\n",
			entry.ID,
			entry.Type,
			entry.CreateTime.Format("2006-01-02 15:04:05"),
			entry.InAmount, entry.InCurrency,
			entry.OutAmount, entry.OutCurrency,
			entry.FeeAmount, entry.FeeCurrency,
		)
	}
	flushTable(w, &buf)
	return nil
}
