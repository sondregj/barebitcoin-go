package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var buyCmd = &cobra.Command{
	Use:   "buy <amount-nok>",
	Short: "Place a market buy order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runBuyCmd(cmd.Context(), client, args)
	},
}

func runBuyCmd(ctx context.Context, client *barebitcoin.HTTPClient, args []string) error {
	amountNOK, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	resp, err := client.CreateOrder(ctx, &barebitcoin.NewOrderRequest{
		Type:      barebitcoin.OrderTypeMarket,
		Direction: barebitcoin.OrderDirectionBuy,
		Amount:    amountNOK,
	})
	if err != nil {
		return err
	}
	fmt.Println("order {")
	fmt.Println("  id", resp.OrderID)
	fmt.Println("  trade id", resp.TradeID)
	fmt.Println("}")
	return nil
}
