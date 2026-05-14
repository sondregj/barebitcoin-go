package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var (
	sendDescription string
	sendIsPayment   bool
)

func init() {
	sendCmd.Flags().StringVar(&sendDescription, "description", "", "Description for the withdrawal")
	sendCmd.Flags().BoolVar(&sendIsPayment, "payment", false, "Mark as a payment (affects tax export)")
}

var sendCmd = &cobra.Command{
	Use:   "send <destination> <amount-btc>",
	Short: "Send bitcoin to an address or Lightning invoice",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runSendCmd(cmd.Context(), client, args, sendDescription, sendIsPayment)
	},
}

func runSendCmd(
	ctx context.Context,
	client *barebitcoin.HTTPClient,
	args []string,
	description string,
	isPayment bool,
) error {
	amountBTC, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	resp, err := client.SendBitcoin(ctx, &barebitcoin.SendBitcoinRequest{
		Destination: args[0],
		AmountBTC:   amountBTC,
		Description: description,
		IsPayment:   isPayment,
	})
	if err != nil {
		return err
	}
	fmt.Println("withdrawal {")
	fmt.Println("  id", resp.WithdrawalID)
	fmt.Println("  network", resp.Network)
	fmt.Println("  status", resp.Status)
	fmt.Println("}")
	return nil
}
