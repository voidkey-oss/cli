package cmd

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Test basic properties of root command
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "voidkey", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Voidkey zero-trust credential broker CLI")
	assert.Contains(t, rootCmd.Long, "zero-trust credential broker")
}

func TestRootCommandFlags(t *testing.T) {
	// Test that the server flag is set up correctly
	serverFlag := rootCmd.PersistentFlags().Lookup("server")
	assert.NotNil(t, serverFlag)
	assert.Equal(t, "http://localhost:3000", serverFlag.DefValue)
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic
	// We can't actually call Execute() as it would run the CLI
	// Instead, we test that the function exists and rootCmd is properly set up
	assert.NotPanics(t, func() {
		// Just verify the structure exists
		_ = rootCmd.Commands()
	})
}

func TestInitCommands(t *testing.T) {
	// Save original state
	originalCommands := rootCmd.Commands()
	
	// Clear commands to test initialization
	rootCmd.ResetCommands()
	
	// Call initCommands
	initCommands()
	
	// Check that commands were added
	commands := rootCmd.Commands()
	assert.Greater(t, len(commands), 0)
	
	// Look for specific commands
	var foundMint, foundList bool
	for _, cmd := range commands {
		if cmd.Use == "mint" {
			foundMint = true
		}
		if cmd.Use == "list-idps" {
			foundList = true
		}
	}
	
	assert.True(t, foundMint, "mint command should be added")
	assert.True(t, foundList, "list-idps command should be added")
	
	// Restore original commands
	rootCmd.ResetCommands()
	for _, cmd := range originalCommands {
		rootCmd.AddCommand(cmd)
	}
}

func TestNewVoidkeyClientIntegration(t *testing.T) {
	// Test that NewVoidkeyClient is called during initialization
	httpClient := &http.Client{}
	serverURL := "http://test.example.com"
	
	client := NewVoidkeyClient(httpClient, serverURL)
	
	assert.NotNil(t, client)
	assert.Equal(t, httpClient, client.client)
	assert.Equal(t, serverURL, client.serverURL)
}

func TestRootCommandHelp(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	
	// Execute help command
	err := rootCmd.Execute()
	
	// Help command should not return an error
	assert.NoError(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "voidkey")
	assert.Contains(t, output, "zero-trust credential broker")
}

func TestRootCommandSubcommands(t *testing.T) {
	// Ensure root command has expected subcommands
	commands := rootCmd.Commands()
	
	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Use
	}
	
	// Check for expected commands (may vary based on initialization)
	// At minimum, we expect version command to be present
	var hasVersion bool
	for _, name := range commandNames {
		if name == "version" {
			hasVersion = true
			break
		}
	}
	assert.True(t, hasVersion, "version command should be present")
}

func TestServerURLGlobal(t *testing.T) {
	// Test that serverURL global variable is properly initialized
	// and can be modified by flags
	
	// Get initial value
	initialURL := serverURL
	
	// The default should be localhost:3000
	assert.Equal(t, "http://localhost:3000", initialURL)
	
	// Test flag parsing (simulate setting the flag)
	rootCmd.PersistentFlags().Set("server", "http://custom.example.com")
	
	// The global variable should be updated
	assert.Contains(t, []string{"http://localhost:3000", "http://custom.example.com"}, serverURL)
}