package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type MintRequest struct {
	OidcToken string `json:"oidcToken"`
	IdpName   string `json:"idpName,omitempty"`
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
	var localIdpName string
	
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint short-lived cloud credentials",
		Long: `Mint short-lived, scoped cloud credentials using OIDC-based authentication.
The credentials are returned as environment variables that can be sourced.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return mintCredentialsWithFlags(voidkeyClient, cobraCmd, localOidcToken, localOutputFormat, localIdpName)
		},
	}

	// Flags for the mint command
	cmd.Flags().StringVar(&localOidcToken, "token", "", "OIDC token for authentication (uses dummy token if not provided)")
	cmd.Flags().StringVarP(&localOutputFormat, "output", "o", "env", "Output format (env|json)")
	cmd.Flags().StringVar(&localIdpName, "idp", "", "IdP provider name to use (uses server default if not specified)")

	return cmd
}

func mintCredentialsWithFlags(client *VoidkeyClient, cmd *cobra.Command, token, format, idpName string) error {
	// Check for token from environment variable if not provided via flag
	if token == "" {
		token = os.Getenv("OIDC_TOKEN")
		if token == "" {
			// Also check for GitHub Actions token as a common case
			token = os.Getenv("GITHUB_TOKEN")
			if token != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using GITHUB_TOKEN environment variable\n")
			}
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using OIDC_TOKEN environment variable\n")
		}
	}

	// Special case for hello-world IdP - provide default token
	if token == "" && idpName == "hello-world" {
		token = "cli-hello-world-token"
		fmt.Fprintf(cmd.ErrOrStderr(), "🎭 Using hello-world IdP with default token\n")
	}

	// Require a valid OIDC token
	if token == "" {
		return fmt.Errorf("OIDC token is required. Provide via:\n" +
			"  --token flag: voidkey mint --token \"your.jwt.token\"\n" +
			"  OIDC_TOKEN env var: export OIDC_TOKEN=\"your.jwt.token\"\n" +
			"  GITHUB_TOKEN env var (for GitHub Actions IdP)\n\n" +
			"To obtain an OIDC token:\n" +
			"  - Auth0: Use the Auth0 CLI or obtain from your application\n" +
			"  - GitHub Actions: Available as ${{ github.token }}\n" +
			"  - Other IdPs: Consult your identity provider's documentation\n" +
			"  - Hello World: Use --idp hello-world for testing (no token required)")
	}

	// Show IdP selection info
	if idpName != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using IdP provider: %s\n", idpName)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using server default IdP provider\n")
	}

	// Mint credentials using client
	credentials, err := client.MintCredentials(token, idpName)
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