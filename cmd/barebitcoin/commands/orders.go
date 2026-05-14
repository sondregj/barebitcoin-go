package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List open orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return runOrdersCmd(cmd.Context(), client)
	},
}

func runOrdersCmd(ctx context.Context, client *barebitcoin.HTTPClient) error {
	resp, err := client.GetOrders(ctx)
	if err != nil {
		return err
	}
	if len(resp.Orders) == 0 {
		fmt.Println("no open orders")
		return nil
	}
	fmt.Println("orders {")
	for _, order := range resp.Orders {
		fmt.Println("  order {")
		fmt.Println("    id", order.OrderID)
		fmt.Println("    type", order.Type)
		fmt.Println("    direction", order.Direction)
		fmt.Println("    amount", order.Amount)
		fmt.Printf("    created %q\n", order.CreatedAt)
		fmt.Println("  }")
	}
	fmt.Println("}")
	return nil
}

var cancelCmd = &cobra.Command{
	Use:   "cancel <order-id>",
	Short: "Cancel an open order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return client.DeleteOrder(cmd.Context(), args[0])
	},
}
