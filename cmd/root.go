package cmd

import (
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	config    *Config

	// clientFactory allows dependency injection for testing
	clientFactory func() *VoidkeyClient = getClient
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "voidkey",
	Short: "Voidkey zero-trust credential broker CLI",
	Long: `Voidkey is a zero-trust credential broker that eliminates long-lived secrets 
in workflows like CI/CD pipelines by dynamically minting short-lived, scoped credentials using OIDC-based authentication.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Load configuration early
	var err error
	config, err = LoadConfig()
	if err != nil {
		// If config loading fails, use defaults but continue
		config = DefaultConfig()
	}

	// Global flags (can override config values)
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", config.Server, "Voidkey broker server URL (overrides config)")

	// Initialize commands after flags are set up
	initCommands()
}

func initCommands() {
	// Initialize commands - client will be created dynamically in each command
	mintCmd := mintCreds(config)
	listIdpsCmd := listIdpProviders()

	rootCmd.AddCommand(mintCmd)
	rootCmd.AddCommand(listIdpsCmd)
	rootCmd.AddCommand(configCmd())
}

// getClient creates a client with the effective server URL (handles flag overrides)
func getClient() *VoidkeyClient {
	return NewVoidkeyClient(&http.Client{}, serverURL)
}
