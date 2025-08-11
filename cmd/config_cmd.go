package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// configCmd creates the config management command
func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long: `Manage the persistent CLI configuration for server URL, IdP provider, and token settings.

Configuration is stored in ~/.voidkey/config.yaml and allows you to set up
common settings once instead of passing them with every command.`,
	}

	cmd.AddCommand(configInitCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configListCmd())

	return cmd
}

// configInitCmd initializes configuration with interactive prompts
func configInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration with interactive setup",
		Long: `Initialize the CLI configuration by prompting for common settings.
This creates ~/.voidkey/config.yaml with your preferences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := DefaultConfig()

			// Prompt for server URL
			fmt.Print("Enter Voidkey broker server URL [http://localhost:3000]: ")
			var serverURL string
			if _, err := fmt.Scanln(&serverURL); err != nil && err.Error() != "unexpected newline" {
				return fmt.Errorf("failed to read server URL: %w", err)
			}
			if strings.TrimSpace(serverURL) != "" {
				config.Server = strings.TrimSpace(serverURL)
			}

			// Prompt for IdP name
			fmt.Print("Enter default IdP provider name (optional): ")
			var idpName string
			if _, err := fmt.Scanln(&idpName); err != nil && err.Error() != "unexpected newline" {
				return fmt.Errorf("failed to read IdP name: %w", err)
			}
			config.IdpName = strings.TrimSpace(idpName)

			// Prompt for token environment variable
			fmt.Print("Enter token environment variable name [OIDC_TOKEN]: ")
			var tokenEnv string
			if _, err := fmt.Scanln(&tokenEnv); err != nil && err.Error() != "unexpected newline" {
				return fmt.Errorf("failed to read token environment variable: %w", err)
			}
			if strings.TrimSpace(tokenEnv) != "" {
				config.TokenEnv = strings.TrimSpace(tokenEnv)
			}

			// Save configuration
			if err := SaveConfig(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			configPath, _ := GetConfigPath()
			fmt.Printf("✅ Configuration saved to %s\n", configPath)
			return nil
		},
	}
}

// configSetCmd sets individual configuration values
func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a specific configuration value.

Available keys:
  server    - Voidkey broker server URL
  idp       - Default IdP provider name
  token-env - Environment variable name for OIDC token

Examples:
  voidkey config set server http://broker.example.com:3000
  voidkey config set idp github-actions
  voidkey config set token-env GITHUB_TOKEN`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			switch key {
			case "server":
				config.Server = value
			case "idp":
				config.IdpName = value
			case "token-env":
				config.TokenEnv = value
			default:
				return fmt.Errorf("unknown configuration key: %s\nValid keys: server, idp, token-env", key)
			}

			if err := SaveConfig(config); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Printf("✅ Set %s = %s\n", key, value)
			return nil
		},
	}
}

// configGetCmd gets individual configuration values
func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get a specific configuration value.

Available keys:
  server    - Voidkey broker server URL
  idp       - Default IdP provider name
  token-env - Environment variable name for OIDC token

Examples:
  voidkey config get server
  voidkey config get idp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			switch key {
			case "server":
				fmt.Println(config.Server)
			case "idp":
				fmt.Println(config.IdpName)
			case "token-env":
				fmt.Println(config.TokenEnv)
			default:
				return fmt.Errorf("unknown configuration key: %s\nValid keys: server, idp, token-env", key)
			}

			return nil
		},
	}
}

// configListCmd lists all configuration values
func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  `List all current configuration values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			configPath, _ := GetConfigPath()
			fmt.Printf("Configuration file: %s\n\n", configPath)
			fmt.Printf("server:    %s\n", config.Server)
			fmt.Printf("idp:       %s\n", config.IdpName)
			fmt.Printf("token-env: %s\n", config.TokenEnv)

			// Show current token status
			token := config.GetToken()
			if token != "" {
				fmt.Printf("\n🔑 Current token: Found (%s)\n", config.TokenEnv)
			} else {
				fmt.Printf("\n❌ Current token: Not found\n")
			}

			return nil
		},
	}
}
