package commands

import (
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

	for _, entry := range entries {
		fmt.Println("entry {")
		fmt.Println("  id", entry.ID)
		fmt.Println("  type", entry.Type)
		fmt.Printf("  created %q\n", entry.CreateTime)
		if entry.InAmount != "" {
			fmt.Println("  in", entry.InAmount, entry.InCurrency)
		}
		if entry.OutAmount != "" {
			fmt.Println("  out", entry.OutAmount, entry.OutCurrency)
		}
		if entry.FeeAmount != "" {
			fmt.Println("  fee", entry.FeeAmount, entry.FeeCurrency)
		}
		fmt.Println("}")
	}
	return nil
}
