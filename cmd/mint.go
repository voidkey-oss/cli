package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type MintRequest struct {
	OidcToken string   `json:"oidcToken"`
	IdpName   string   `json:"idpName,omitempty"`
	Keyset    string   `json:"keyset,omitempty"`    // Legacy - for backward compatibility
	Keys      []string `json:"keys,omitempty"`      // New key-based approach
	Duration  int      `json:"duration,omitempty"`  // Optional duration override
}

type CloudCredentials struct {
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	SessionToken string `json:"sessionToken,omitempty"`
	ExpiresAt    string `json:"expiresAt"`
}

// New key-based response
type MultiKeyResponse struct {
	Keys map[string]KeyCredentialResponse `json:"keys"`
}

// mintCreds creates a new mint command with dependency injection
func mintCreds(voidkeyClient *VoidkeyClient) *cobra.Command {
	var localOidcToken string
	var localOutputFormat string
	var localIdpName string
	var localKeyset string
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
  voidkey mint --keys MINIO_CREDENTIALS --duration 1800

  # Legacy keyset usage (backward compatibility)
  voidkey mint --keyset dev`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return mintCredentialsWithFlags(voidkeyClient, cobraCmd, localOidcToken, localOutputFormat, localIdpName, localKeyset, localKeys, localDuration, localAll)
		},
	}

	// Flags for the mint command
	cmd.Flags().StringVar(&localOidcToken, "token", "", "OIDC token for authentication (uses dummy token if not provided)")
	cmd.Flags().StringVarP(&localOutputFormat, "output", "o", "env", "Output format (env|json)")
	cmd.Flags().StringVar(&localIdpName, "idp", "", "IdP provider name to use (uses server default if not specified)")
	cmd.Flags().StringVar(&localKeyset, "keyset", "", "[LEGACY] Keyset name to use for environment variable mapping")
	cmd.Flags().StringSliceVar(&localKeys, "keys", nil, "Comma-separated list of key names to mint (e.g. MINIO_CREDENTIALS,AWS_CREDENTIALS)")
	cmd.Flags().IntVar(&localDuration, "duration", 0, "Duration in seconds to override default credential lifetime")
	cmd.Flags().BoolVar(&localAll, "all", false, "Mint all available keys for the identity")

	return cmd
}

func mintCredentialsWithFlags(client *VoidkeyClient, cmd *cobra.Command, token, format, idpName, keyset string, keys []string, duration int, all bool) error {
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

	// Determine approach: new key-based vs legacy keyset
	if len(keys) > 0 || all {
		// New key-based approach
		if all {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Minting all available keys\n")
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 Minting keys: %v\n", keys)
		}
		
		if duration > 0 {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⏱️ Duration override: %d seconds\n", duration)
		}

		// Use new key-based minting
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
	} else {
		// Legacy keyset approach for backward compatibility
		if keyset != "" {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 [LEGACY] Using keyset: %s\n", keyset)
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "🔑 [LEGACY] Using all available keys from identity keysets\n")
		}

		// Mint credentials using legacy client method
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

// New output functions for key-based approach
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
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "💡 To use: eval \"$(voidkey mint --keys MINIO_CREDENTIALS)\"\n")
}

func outputKeysAsJSON(keyResponses map[string]KeyCredentialResponse, cmd *cobra.Command) {
	output, _ := json.MarshalIndent(keyResponses, "", "  ")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(output))
}

// init function removed - commands are now initialized in root.go
