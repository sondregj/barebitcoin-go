package commands

import (
	"bytes"
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

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "NETWORK\tDESTINATION")
	if resp.OnchainAddress != nil {
		fmt.Fprintf(w, "onchain\t%s\n", resp.OnchainAddress.Destination)
	}
	if resp.LightningAddress != nil {
		fmt.Fprintf(w, "lightning\t%s\n", resp.LightningAddress.Destination)
	}
	if resp.LNURLPay != nil {
		fmt.Fprintf(w, "lnurl\t%s\n", resp.LNURLPay.Destination)
	}
	flushTable(w, &buf)
	return nil
}
