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
	Keyset    string `json:"keyset,omitempty"`
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
	var localKeyset string

	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint short-lived cloud credentials",
		Long: `Mint short-lived, scoped cloud credentials using OIDC-based authentication.
The credentials are returned as environment variables that can be sourced.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return mintCredentialsWithFlags(voidkeyClient, cobraCmd, localOidcToken, localOutputFormat, localIdpName, localKeyset)
		},
	}

	// Flags for the mint command
	cmd.Flags().StringVar(&localOidcToken, "token", "", "OIDC token for authentication (uses dummy token if not provided)")
	cmd.Flags().StringVarP(&localOutputFormat, "output", "o", "env", "Output format (env|json)")
	cmd.Flags().StringVar(&localIdpName, "idp", "", "IdP provider name to use (uses server default if not specified)")
	cmd.Flags().StringVar(&localKeyset, "keyset", "", "Keyset name to use for environment variable mapping (uses all available keys if not specified)")

	return cmd
}

func mintCredentialsWithFlags(client *VoidkeyClient, cmd *cobra.Command, token, format, idpName, keyset string) error {
	// Check for token from environment variable if not provided via flag
	if token == "" {
		token = os.Getenv("OIDC_TOKEN")
		if token == "" {
			// Also check for GitHub Actions token as a common case
			token = os.Getenv("GITHUB_TOKEN")
			if token != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using GITHUB_TOKEN environment variable\n")
			}
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using OIDC_TOKEN environment variable\n")
		}
	}

	// Special case for hello-world IdP - provide default token
	if token == "" && idpName == "hello-world" {
		token = "cli-hello-world-token"
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🎭 Using hello-world IdP with default token\n")
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
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using IdP provider: %s\n", idpName)
	} else {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Using server default IdP provider\n")
	}

	// Show keyset selection info
	if keyset != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Using keyset: %s\n", keyset)
	} else {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Using all available keys from identity keysets\n")
	}

	// Mint credentials using client
	credentials, err := client.MintCredentials(token, idpName, keyset)
	if err != nil {
		return err
	}

	// If a specific keyset was requested, also get the environment variable mappings
	var keysetVars map[string]string
	if keyset != "" {
		keysetVars, err = client.GetKeysetKeysWithToken(token, keyset)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Warning: Could not retrieve keyset environment variables: %v\n", err)
		}
	}

	// Output credentials in requested format
	switch format {
	case "env":
		outputAsEnvVars(*credentials, keysetVars, cmd)
	case "json":
		outputAsJSON(*credentials, cmd)
	default:
		outputAsEnvVars(*credentials, keysetVars, cmd) // default format
	}

	return nil
}

func outputAsEnvVars(creds CloudCredentials, keysetVars map[string]string, cmd *cobra.Command) {
	// Always output the core AWS credentials
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKey)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretKey)
	if creds.SessionToken != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export AWS_CREDENTIAL_EXPIRATION=%s\n", creds.ExpiresAt)

	// If keyset variables are available, output them as well
	if keysetVars != nil && len(keysetVars) > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Setting keyset environment variables:\n")
		for envVar, value := range keysetVars {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export %s=%s\n", envVar, value)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  %s=%s\n", envVar, value)
		}
	}

	// Print success message to stderr so it doesn't interfere with sourcing
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "✅ Credentials minted successfully (expires: %s)\n", creds.ExpiresAt)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "💡 To use: eval \"$(voidkey mint)\"\n")
}

func outputAsJSON(creds CloudCredentials, cmd *cobra.Command) {
	output, _ := json.MarshalIndent(creds, "", "  ")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(output))
}

// init function removed - commands are now initialized in root.go
