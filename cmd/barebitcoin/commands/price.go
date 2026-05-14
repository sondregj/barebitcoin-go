package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "Fetch the current BTCNOK price",
	RunE:  runPriceCmd,
}

func runPriceCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	client := barebitcoin.NewHTTPClient()

	price, err := client.GetPrice(ctx, 0)
	if err != nil {
		return err
	}
	fmt.Println("price {")
	fmt.Println("  mid btc nok", strconv.FormatFloat(price.Price, 'f', -1, 64))
	fmt.Println("  ask btc nok", strconv.FormatFloat(price.Ask, 'f', -1, 64))
	fmt.Println("  bid btc nok", strconv.FormatFloat(price.Bid, 'f', -1, 64))
	fmt.Printf("  timestamp %q\n", price.Timestamp)
	fmt.Println("}")
	return nil
}
