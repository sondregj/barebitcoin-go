package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var invoiceCmd = &cobra.Command{
	Use:   "invoice <satoshi>",
	Short: "Create a Lightning invoice",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runInvoiceCmd(cmd.Context(), client, args)
	},
}

func runInvoiceCmd(ctx context.Context, client *barebitcoin.HTTPClient, args []string) error {
	amountSatoshi, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	amountBTC := float64(amountSatoshi) / 1e8
	invoice, err := client.CreateLightningInvoice(ctx, &barebitcoin.NewLightningInvoiceRequest{
		Currency: barebitcoin.CurrencyBTC,
		Amount:   amountBTC,
	})
	if err != nil {
		return err
	}
	fmt.Println(invoice.Invoice)
	return nil
}

func init() {
	invoiceCmd.AddCommand(invoiceStatusCmd)
}

var invoiceStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Fetch the status of a Lightning invoice",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runInvoiceStatusCmd(cmd.Context(), client, args[0])
	},
}

func runInvoiceStatusCmd(ctx context.Context, client *barebitcoin.HTTPClient, id string) error {
	invoice, err := client.GetLightningInvoice(ctx, id)
	if err != nil {
		return err
	}
	fmt.Println("invoice {")
	fmt.Println("  id", invoice.DepositDestinationID)
	fmt.Println("  status", invoice.Status)
	fmt.Println("  invoice", invoice.Invoice)
	fmt.Println("}")
	return nil
}
