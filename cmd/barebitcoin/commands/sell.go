package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var sellCmd = &cobra.Command{
	Use:   "sell <amount-btc>",
	Short: "Place a market sell order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runSellCmd(cmd.Context(), client, args)
	},
}

func runSellCmd(ctx context.Context, client *barebitcoin.HTTPClient, args []string) error {
	amountBTC, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	resp, err := client.CreateOrder(ctx, &barebitcoin.NewOrderRequest{
		Type:      barebitcoin.OrderTypeMarket,
		Direction: barebitcoin.OrderDirectionSell,
		Amount:    amountBTC,
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
