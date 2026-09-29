package commands

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

var volumeDate string

func init() {
	volumeCmd.PersistentFlags().StringVar(&volumeDate, "date", "", "Date to fetch volume for (YYYY-MM-DD), defaults to today")
	volumeCmd.AddCommand(volumeHistoricCmd)
}

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Fetch BTCNOK trade volume statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Does not require authentication
		client := barebitcoin.NewHTTPClientWithKeys("", "")
		return runVolumeCmd(cmd.Context(), client, volumeDate)
	},
}

func runVolumeCmd(ctx context.Context, client *barebitcoin.HTTPClient, date string) error {
	resp, err := client.GetVolume(ctx, date)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "PERIOD\tNOK\tBTC\tBUY %\tTRADES")
	for _, period := range []struct {
		name  string
		stats barebitcoin.VolumeStats
	}{
		{"daily", resp.Daily},
		{"monthly", resp.Monthly},
		{"yearly", resp.Yearly},
		{"last 24h", resp.Last24h},
	} {
		fmt.Fprintf(w, "%s\t%s\t%s\t%g\t%d\n",
			period.name,
			formatFloat(period.stats.AmountNOK),
			formatFloat(period.stats.AmountBTC),
			period.stats.BuyPercentage,
			period.stats.NumberOfTrades,
		)
	}
	flushTable(w, &buf)
	return nil
}

var volumeHistoricCmd = &cobra.Command{
	Use:   "historic",
	Short: "Fetch historic daily volume per market",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Does not require authentication
		client := barebitcoin.NewHTTPClientWithKeys("", "")
		return runVolumeHistoricCmd(cmd.Context(), client, volumeDate)
	},
}

func runVolumeHistoricCmd(ctx context.Context, client *barebitcoin.HTTPClient, date string) error {
	resp, err := client.GetVolumeHistoric(ctx, date)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := newTabWriter(&buf)
	fmt.Fprintln(w, "DATE\tMARKET\tNOK\tBTC\tSHARE")
	for _, day := range resp.DailyVolume {
		for _, key := range slices.Sorted(maps.Keys(day.Stats)) {
			stats := day.Stats[key]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.1f%%\n",
				day.Date,
				stats.Name,
				formatFloat(stats.VolumeNOK),
				formatFloat(stats.VolumeBTC),
				stats.ShareOfTotalVolume*100,
			)
		}
	}
	flushTable(w, &buf)

	if len(resp.ShareOfTotalVolume) > 0 {
		fmt.Println()
		buf.Reset()
		w = newTabWriter(&buf)
		fmt.Fprintln(w, "MARKET\tSHARE OF TOTAL")
		for _, market := range slices.Sorted(maps.Keys(resp.ShareOfTotalVolume)) {
			fmt.Fprintf(w, "%s\t%.1f%%\n", market, resp.ShareOfTotalVolume[market]*100)
		}
		flushTable(w, &buf)
	}
	return nil
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
