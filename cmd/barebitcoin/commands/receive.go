package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var receiveAccountID string

func init() {
	receiveCmd.Flags().StringVar(&receiveAccountID, "account-id", "", "Account to get addresses for")
}

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Show bitcoin deposit addresses",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReceiveCmd(cmd.Context(), receiveAccountID)
	},
}

func runReceiveCmd(ctx context.Context, accountID string) error {
	resp, err := barebitcoin.ListBitcoinDepositDestinations(ctx, accountID)
	if err != nil {
		return err
	}
	fmt.Println("destinations {")
	if resp.OnchainAddress != nil {
		fmt.Println("  onchain", resp.OnchainAddress.Destination)
	}
	if resp.LightningAddress != nil {
		fmt.Println("  lightning address", resp.LightningAddress.Destination)
	}
	if resp.LNURLPay != nil {
		fmt.Println("  lnurl", resp.LNURLPay.Destination)
	}
	fmt.Println("}")
	return nil
}
