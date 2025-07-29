package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type MintKeysRequest struct {
	OidcToken string   `json:"oidcToken"`
	IdpName   string   `json:"idpName,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	Duration  int      `json:"duration,omitempty"`
	All       bool     `json:"all,omitempty"`
}

// mintCreds creates a new mint command with dependency injection
func mintCreds(voidkeyClient *VoidkeyClient) *cobra.Command {
	var localOidcToken string
	var localOutputFormat string
	var localIdpName string
	var localKeys []string
	var localDuration int
	var localAll bool

	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Mint short-lived cloud credentials",
		Long: `Mint short-lived, scoped cloud credentials using OIDC-based authentication.
The credentials are returned as environment variables that can be sourced.

Examples:
  # Mint specific keys
  voidkey mint --keys MINIO_CREDENTIALS,AWS_CREDENTIALS

  # Mint all available keys
  voidkey mint --all

  # Mint with custom duration (in seconds)
  voidkey mint --keys MINIO_CREDENTIALS --duration 1800`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return mintCredentialsWithFlags(voidkeyClient, cobraCmd, localOidcToken, localOutputFormat, localIdpName, localKeys, localDuration, localAll)
		},
	}

	// Flags for the mint command
	cmd.Flags().StringVar(&localOidcToken, "token", "", "OIDC token for authentication (uses dummy token if not provided)")
	cmd.Flags().StringVarP(&localOutputFormat, "output", "o", "env", "Output format (env|json)")
	cmd.Flags().StringVar(&localIdpName, "idp", "", "IdP provider name to use (uses server default if not specified)")
	cmd.Flags().StringSliceVar(&localKeys, "keys", nil, "Comma-separated list of key names to mint (e.g. MINIO_CREDENTIALS,AWS_CREDENTIALS)")
	cmd.Flags().IntVar(&localDuration, "duration", 0, "Duration in seconds to override default credential lifetime")
	cmd.Flags().BoolVar(&localAll, "all", false, "Mint all available keys for the identity")

	return cmd
}

func mintCredentialsWithFlags(client *VoidkeyClient, cmd *cobra.Command, token, format, idpName string, keys []string, duration int, all bool) error {
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

	// Validate that at least one approach is specified
	if len(keys) == 0 && !all {
		return fmt.Errorf("must specify either specific keys (--keys) or all keys (--all)")
	}

	// Use key-based minting
	if all {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Minting all available keys\n")
	} else {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Minting keys: %v\n", keys)
	}

	if duration > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⏱️ Duration override: %d seconds\n", duration)
	}

	keyResponses, err := client.MintKeys(token, idpName, keys, duration, all)
	if err != nil {
		return err
	}

	// Output credentials in requested format
	switch format {
	case "env":
		outputKeysAsEnvVars(keyResponses, cmd)
	case "json":
		outputKeysAsJSON(keyResponses, cmd)
	default:
		outputKeysAsEnvVars(keyResponses, cmd) // default format
	}

	return nil
}

// Key-based output functions
func outputKeysAsEnvVars(keyResponses map[string]KeyCredentialResponse, cmd *cobra.Command) {
	totalVars := 0

	for keyName, response := range keyResponses {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Key: %s (expires: %s)\n", keyName, response.ExpiresAt)

		for envVar, value := range response.Credentials {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export %s=%s\n", envVar, value)
			totalVars++
		}
	}

	// Print success message to stderr so it doesn't interfere with sourcing
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "✅ Successfully minted %d keys with %d environment variables\n", len(keyResponses), totalVars)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "💡 To use: eval \"$(voidkey mint --all)\" or eval \"$(voidkey mint --keys KEY_NAME)\"\n")
}

func outputKeysAsJSON(keyResponses map[string]KeyCredentialResponse, cmd *cobra.Command) {
	output, _ := json.MarshalIndent(keyResponses, "", "  ")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(output))
}

// init function removed - commands are now initialized in root.go
