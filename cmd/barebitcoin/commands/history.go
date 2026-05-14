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
	Short: "Fetch ledger history",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runHistoryCmd(cmd.Context(), client, historyLimit)
	},
}

func runHistoryCmd(ctx context.Context, client *barebitcoin.HTTPClient, limit int) error {
	ledger, err := client.GetLedger(ctx)
	if err != nil {
		return err
	}

	entries := ledger.Entries
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	for _, entry := range entries {
		fmt.Println("entry {")
		fmt.Println("  id", entry.TransactionID)
		fmt.Println("  type", entry.Type)
		fmt.Printf("  timestamp %q\n", entry.Timestamp)
		fmt.Println("  currency", entry.Currency)
		fmt.Println("  value", entry.Value)
		fmt.Println("  fee", entry.Fee)
		fmt.Println("  balance", entry.Balance)
		fmt.Println("}")
	}
	return nil
}
