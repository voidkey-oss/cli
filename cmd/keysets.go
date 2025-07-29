package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// keysetsCmd creates a new keysets command with dependency injection
func keysetsCmd(voidkeyClient *VoidkeyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keysets",
		Short: "Manage keysets for identities",
		Long:  `List and manage keysets that define environment variable mappings for identities.`,
	}

	// Add subcommands
	cmd.AddCommand(listKeysetsCmd(voidkeyClient))
	cmd.AddCommand(showKeysetCmd(voidkeyClient))

	return cmd
}

// listKeysetsCmd lists all available keysets for a subject
func listKeysetsCmd(voidkeyClient *VoidkeyClient) *cobra.Command {
	var subject string
	var token string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available keysets for a subject",
		Long:  `List all available keysets and their environment variable mappings for a given subject or token.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			var keysets map[string]map[string]string
			var err error

			if token != "" {
				keysets, err = voidkeyClient.GetAvailableKeysetsWithToken(token)
				if err != nil {
					return err
				}
			} else if subject != "" {
				keysets, err = voidkeyClient.GetAvailableKeysets(subject)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("either subject or token is required. Use --subject or --token flag")
			}

			if keysets == nil {
				if token != "" {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "No keysets found for the provided token\n")
				} else {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "No keysets found for subject: %s\n", subject)
				}
				return nil
			}

			switch outputFormat {
			case "json":
				output, _ := json.MarshalIndent(keysets, "", "  ")
				_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "%s\n", string(output))
			default:
				if token != "" {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Available keysets for the provided token:\n\n")
				} else {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Available keysets for subject '%s':\n\n", subject)
				}
				for keysetName, keys := range keysets {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "📦 %s:\n", keysetName)
					for envVar, value := range keys {
						_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "  %s=%s\n", envVar, value)
					}
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Subject (identity) to query keysets for")
	cmd.Flags().StringVar(&token, "token", "", "OIDC token to extract subject from and query keysets")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json)")

	return cmd
}

// showKeysetCmd shows keys for a specific keyset
func showKeysetCmd(voidkeyClient *VoidkeyClient) *cobra.Command {
	var subject string
	var token string
	var keysetName string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show keys for a specific keyset",
		Long:  `Show the environment variable mappings for a specific keyset of a given subject or token.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if keysetName == "" {
				return fmt.Errorf("keyset is required. Use --keyset flag to specify the keyset name")
			}

			var keys map[string]string
			var err error

			if token != "" {
				keys, err = voidkeyClient.GetKeysetKeysWithToken(token, keysetName)
				if err != nil {
					return err
				}
			} else if subject != "" {
				keys, err = voidkeyClient.GetKeysetKeys(subject, keysetName)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("either subject or token is required. Use --subject or --token flag")
			}

			if keys == nil {
				if token != "" {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Keyset '%s' not found for the provided token\n", keysetName)
				} else {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Keyset '%s' not found for subject: %s\n", keysetName, subject)
				}
				return nil
			}

			switch outputFormat {
			case "json":
				output, _ := json.MarshalIndent(keys, "", "  ")
				_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "%s\n", string(output))
			case "env":
				for envVar, value := range keys {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "export %s=%s\n", envVar, value)
				}
			default:
				if token != "" {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Keys in keyset '%s' for the provided token:\n\n", keysetName)
				} else {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "Keys in keyset '%s' for subject '%s':\n\n", keysetName, subject)
				}
				for envVar, value := range keys {
					_, _ = fmt.Fprintf(cobraCmd.OutOrStdout(), "%s=%s\n", envVar, value)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Subject (identity) to query")
	cmd.Flags().StringVar(&token, "token", "", "OIDC token to extract subject from and query")
	cmd.Flags().StringVar(&keysetName, "keyset", "", "Keyset name to show")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|env)")
	cmd.MarkFlagRequired("keyset")

	return cmd
}
