package cmd

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListIdpProviders_CommandCreation(t *testing.T) {
	cmd := listIdpProviders()

	assert.NotNil(t, cmd)
	assert.Equal(t, "list-idps", cmd.Use)
	assert.Contains(t, cmd.Short, "List available Identity Providers")
}

func TestListIdpProviders_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: true},
		{Name: "github", IsDefault: false},
		{Name: "hello-world", IsDefault: false},
	}

	MockSuccessfulListResponse(mockClient, "http://localhost:3000", expectedProviders)

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DEFAULT")
	assert.Contains(t, output, "auth0")
	assert.Contains(t, output, "github")
	assert.Contains(t, output, "hello-world")
	assert.Contains(t, output, "✓") // Default indicator for auth0

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_EmptyList(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	emptyProviders := []IdpProvider{}
	MockSuccessfulListResponse(mockClient, "http://localhost:3000", emptyProviders)

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "No Identity Providers configured")

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_SingleProvider(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: true},
	}

	MockSuccessfulListResponse(mockClient, "http://localhost:3000", expectedProviders)

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "auth0")
	assert.Contains(t, output, "✓")

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_MultipleProvidersNoDefault(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: false},
		{Name: "github", IsDefault: false},
		{Name: "okta", IsDefault: false},
	}

	MockSuccessfulListResponse(mockClient, "http://localhost:3000", expectedProviders)

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "auth0")
	assert.Contains(t, output, "github")
	assert.Contains(t, output, "okta")

	// Count occurrences of ✓ (should be 0)
	checkmarkCount := strings.Count(output, "✓")
	assert.Equal(t, 0, checkmarkCount)

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	MockErrorResponse(mockClient, "GET", "http://localhost:3000/credentials/idp-providers", http.StatusInternalServerError, "Internal server error")

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list IdP providers")

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_TableFormatting(t *testing.T) {
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	expectedProviders := []IdpProvider{
		{Name: "very-long-provider-name", IsDefault: true},
		{Name: "short", IsDefault: false},
	}

	MockSuccessfulListResponse(mockClient, "http://localhost:3000", expectedProviders)

	cmd := listIdpProviders()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header, separator, and 2 data rows
	assert.GreaterOrEqual(t, len(lines), 4)

	// Check header
	assert.Contains(t, lines[0], "NAME")
	assert.Contains(t, lines[0], "DEFAULT")

	// Check separator
	assert.Contains(t, lines[1], "----")
	assert.Contains(t, lines[1], "-------")

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_Integration(t *testing.T) {
	// This test ensures the command can be executed as part of a cobra command tree
	mockClient := &MockHTTPClient{}
	cleanup := SetTestClientFactory(mockClient, "http://localhost:3000")
	defer cleanup()

	expectedProviders := []IdpProvider{
		{Name: "test-provider", IsDefault: true},
	}

	MockSuccessfulListResponse(mockClient, "http://localhost:3000", expectedProviders)

	// Create a root command and add the list command
	rootCmd := &cobra.Command{Use: "voidkey"}
	listCmd := listIdpProviders()
	rootCmd.AddCommand(listCmd)

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"list-idps"})

	err := rootCmd.Execute()

	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "test-provider")

	mockClient.AssertExpectations(t)
}
