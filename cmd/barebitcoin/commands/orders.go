package commands

import (
	"bytes"
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

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "ID\tTYPE\tDIRECTION\tAMOUNT\tCREATED")
	for _, order := range resp.Orders {
		fmt.Fprintf(w, "%s\t%s\t%s\t%g\t%s\n",
			order.OrderID,
			order.Type,
			order.Direction,
			order.Amount,
			order.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	flushTable(w, &buf)
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
