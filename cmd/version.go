package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionInfo = struct {
		version string
		commit  string
		date    string
	}{
		version: "dev",
		commit:  "none",
		date:    "unknown",
	}
)

// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, commit, and build date information for the voidkey CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "voidkey version %s\n", versionInfo.version)
		fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", versionInfo.commit)
		fmt.Fprintf(cmd.OutOrStdout(), "built: %s\n", versionInfo.date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}