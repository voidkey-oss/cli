package cmd

import (
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
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
	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:3000", "Voidkey broker server URL")

	// Initialize commands after flags are set up
	initCommands()
}

func initCommands() {
	// Initialize client
	client := NewVoidkeyClient(&http.Client{}, serverURL)

	// Initialize commands with dependency injection
	mintCmd := mintCreds(client)
	listIdpsCmd := listIdpProviders(client)
	keysetsCmd := keysetsCmd(client)

	rootCmd.AddCommand(mintCmd)
	rootCmd.AddCommand(listIdpsCmd)
	rootCmd.AddCommand(keysetsCmd)
}
