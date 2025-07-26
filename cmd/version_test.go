package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSetVersionInfo(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
	}{
		{
			name:    "valid version info",
			version: "1.0.0",
			commit:  "abc123",
			date:    "2024-01-01",
		},
		{
			name:    "empty strings",
			version: "",
			commit:  "",
			date:    "",
		},
		{
			name:    "special characters",
			version: "v1.0.0-beta+build.1",
			commit:  "abc123def456",
			date:    "2024-01-01T10:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersionInfo(tt.version, tt.commit, tt.date)
			
			if versionInfo.version != tt.version {
				t.Errorf("Expected version %s, got %s", tt.version, versionInfo.version)
			}
			if versionInfo.commit != tt.commit {
				t.Errorf("Expected commit %s, got %s", tt.commit, versionInfo.commit)
			}
			if versionInfo.date != tt.date {
				t.Errorf("Expected date %s, got %s", tt.date, versionInfo.date)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
	}{
		{
			name:    "default version info",
			version: "dev",
			commit:  "none",
			date:    "unknown",
		},
		{
			name:    "custom version info",
			version: "1.0.0",
			commit:  "abc123",
			date:    "2024-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set version info
			SetVersionInfo(tt.version, tt.commit, tt.date)

			// Create a buffer to capture output
			var buf bytes.Buffer
			
			// Create a new version command for testing
			cmd := &cobra.Command{
				Use:   "version",
				Short: "Show version information",
				Run: func(cmd *cobra.Command, args []string) {
					cmd.SetOut(&buf)
					versionCmd.Run(cmd, args)
				},
			}
			cmd.SetOut(&buf)

			// Execute the command
			versionCmd.SetOut(&buf)
			versionCmd.Run(cmd, []string{})

			output := buf.String()

			// Check that output contains expected version info
			if !strings.Contains(output, tt.version) {
				t.Errorf("Expected output to contain version %s, got: %s", tt.version, output)
			}
			
			if !strings.Contains(output, tt.commit) {
				t.Errorf("Expected output to contain commit %s, got: %s", tt.commit, output)
			}
			
			if !strings.Contains(output, tt.date) {
				t.Errorf("Expected output to contain date %s, got: %s", tt.date, output)
			}
		})
	}
}

func TestVersionCommandExecution(t *testing.T) {
	// Test that version command can be executed without errors
	var buf bytes.Buffer
	
	// Set version info for testing
	SetVersionInfo("1.0.0", "abc123", "2024-01-01")
	
	// Create a test command with the same output
	testCmd := &cobra.Command{}
	testCmd.SetOut(&buf)
	
	// Execute the version command's Run function directly
	versionCmd.Run(testCmd, []string{})

	output := buf.String()
	if output == "" {
		t.Error("Version command produced no output")
	}

	// Should contain the word "version"
	if !strings.Contains(output, "version") {
		t.Errorf("Expected output to contain 'version', got: %s", output)
	}
}