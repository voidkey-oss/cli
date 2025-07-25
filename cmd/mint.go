package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type MintRequest struct {
	OidcToken string `json:"oidcToken"`
}

type CloudCredentials struct {
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	SessionToken string `json:"sessionToken,omitempty"`
	ExpiresAt    string `json:"expiresAt"`
}

// mintCreds creates a new mint command with dependency injection
func mintCreds(voidkeyClient *VoidkeyClient) *cobra.Command {
	var localOidcToken string
	var localOutputFormat string
	
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint short-lived cloud credentials",
		Long: `Mint short-lived, scoped cloud credentials using OIDC-based authentication.
The credentials are returned as environment variables that can be sourced.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return mintCredentialsWithFlags(voidkeyClient, cobraCmd, localOidcToken, localOutputFormat)
		},
	}

	// Flags for the mint command
	cmd.Flags().StringVar(&localOidcToken, "token", "", "OIDC token for authentication (uses dummy token if not provided)")
	cmd.Flags().StringVarP(&localOutputFormat, "output", "o", "env", "Output format (env|json)")

	return cmd
}

func mintCredentialsWithFlags(client *VoidkeyClient, cmd *cobra.Command, token, format string) error {
	// Use dummy token for hello world if none provided
	if token == "" {
		token = "cli-hello-world-token"
		fmt.Fprintf(cmd.ErrOrStderr(), "🔄 Using dummy OIDC token for hello world demo\n")
	}

	// Mint credentials using client
	credentials, err := client.MintCredentials(token)
	if err != nil {
		return err
	}

	// Output credentials in requested format
	switch format {
	case "env":
		outputAsEnvVars(*credentials, cmd)
	case "json":
		outputAsJSON(*credentials, cmd)
	default:
		outputAsEnvVars(*credentials, cmd) // default format
	}

	return nil
}

func outputAsEnvVars(creds CloudCredentials, cmd *cobra.Command) {
	fmt.Fprintf(cmd.OutOrStdout(), "export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKey)
	fmt.Fprintf(cmd.OutOrStdout(), "export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretKey)
	if creds.SessionToken != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "export AWS_CREDENTIAL_EXPIRATION=%s\n", creds.ExpiresAt)
	
	// Print success message to stderr so it doesn't interfere with sourcing
	fmt.Fprintf(cmd.ErrOrStderr(), "✅ Credentials minted successfully (expires: %s)\n", creds.ExpiresAt)
	fmt.Fprintf(cmd.ErrOrStderr(), "💡 To use: eval \"$(voidkey mint)\"\n")
}

func outputAsJSON(creds CloudCredentials, cmd *cobra.Command) {
	output, _ := json.MarshalIndent(creds, "", "  ")
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(output))
}

// init function removed - commands are now initialized in root.go