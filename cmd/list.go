package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// listIdpProviders creates a new list command with dependency injection
func listIdpProviders(voidkeyClient *VoidkeyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-idps",
		Short: "List available Identity Providers",
		Long:  `List all configured Identity Providers available in the broker server, including which one is set as default.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get providers from server
			providers, err := voidkeyClient.ListIdpProviders()
			if err != nil {
				return fmt.Errorf("failed to list IdP providers: %w", err)
			}

			if len(providers) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No Identity Providers configured")
				return nil
			}

			// Create table writer for nice formatting using command's output
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tDEFAULT")
			_, _ = fmt.Fprintln(w, "----\t-------")

			for _, provider := range providers {
				defaultIndicator := ""
				if provider.IsDefault {
					defaultIndicator = "✓"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\n", provider.Name, defaultIndicator)
			}

			// Flush the table
			if err := w.Flush(); err != nil {
				return fmt.Errorf("failed to display table: %w", err)
			}

			return nil
		},
	}

	return cmd
}
