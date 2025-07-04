package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sondregj/barebitcoin-go"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("Usage: barebitcoin <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	// bitcoin := "₿"
	// bitcoinSatoshi := "₿̰"

	client := barebitcoin.NewHTTPClient()

	switch command {

	case "price":
		price, err := client.GetPrice(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("price {")
		fmt.Println("  mid btc nok", strconv.FormatFloat(price.Price, 'f', -1, 64))
		fmt.Println("  ask btc nok", strconv.FormatFloat(price.Ask, 'f', -1, 64))
		fmt.Println("  bid btc nok", strconv.FormatFloat(price.Bid, 'f', -1, 64))
		fmt.Printf("  timestamp %q\n", price.Timestamp)
		fmt.Println("}")

	case "user":
		// TODO: user info
		fmt.Println("error: user info not implemented")
		os.Exit(1)

	case "holdings":
		user, err := client.GetBitcoinAccounts(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var totalAvailableBTC, totalPendingOrdersBTC float64
		fmt.Println("accounts {")
		for _, account := range user.Accounts {
			fmt.Println("  account {")
			fmt.Println("    id", account.ID)
			fmt.Println("    available bitcoin", account.AvailableBTC)
			fmt.Println("    pending orders bitcoin", account.PendingOrdersBTC)
			fmt.Println("  }")
			totalAvailableBTC += account.AvailableBTC
			totalPendingOrdersBTC += account.PendingOrdersBTC
		}
		fmt.Println("}")
		fmt.Println("total available bitcoin", totalAvailableBTC)
		fmt.Println("total pending orders bitcoin", totalPendingOrdersBTC)

	case "stats":
		fmt.Println("error: stats not implemented")
		os.Exit(1)

	case "history":
		fmt.Println("error: history not implemented")
		os.Exit(1)

	case "invoice":
		var amountSatoshi int
		if len(os.Args) == 3 {
			var err error
			amountSatoshi, err = strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		amountBTC := float64(amountSatoshi) / 1e8
		invoice, err := client.CreateLightningInvoice(ctx, &barebitcoin.NewLightningInvoiceRequest{
			Currency: barebitcoin.CurrencyBTC,
			Amount:   amountBTC,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(invoice.Invoice)

	default:
		fmt.Printf("unknown command %q\n", command)
	}
}
