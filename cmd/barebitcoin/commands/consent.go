package commands

import (
	"github.com/spf13/cobra"

	"github.com/sondregj/barebitcoin-go"
)

func init() {
	consentCmd.AddCommand(consentRevokeCmd)
}

var consentCmd = &cobra.Command{
	Use:   "consent",
	Short: "Manage consent given to OAuth2 applications",
}

var consentRevokeCmd = &cobra.Command{
	Use:   "revoke <client-id>",
	Short: "Revoke consent for an OAuth2 application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := barebitcoin.NewHTTPClient()
		return client.RevokeConsent(cmd.Context(), args[0])
	},
}
