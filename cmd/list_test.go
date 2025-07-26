package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListIdpProviders_CommandCreation(t *testing.T) {
	mockClient := &MockHTTPClient{}
	voidkeyClient := NewVoidkeyClient(mockClient, "http://localhost:3000")

	cmd := listIdpProviders(voidkeyClient)

	assert.NotNil(t, cmd)
	assert.Equal(t, "list-idps", cmd.Use)
	assert.Contains(t, cmd.Short, "List available Identity Providers")
}

func TestListIdpProviders_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: true},
		{Name: "github", IsDefault: false},
		{Name: "hello-world", IsDefault: false},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
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
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	responseBody, _ := json.Marshal([]IdpProvider{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
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
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: true},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
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
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: false},
		{Name: "github", IsDefault: false},
		{Name: "okta", IsDefault: false},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
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
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Internal server error")),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.RunE(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list IdP providers")

	mockClient.AssertExpectations(t)
}

func TestListIdpProviders_TableFormatting(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "very-long-provider-name", IsDefault: true},
		{Name: "short", IsDefault: false},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	cmd := listIdpProviders(client)
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
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "test-provider", IsDefault: true},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	// Create a root command and add the list command
	rootCmd := &cobra.Command{Use: "voidkey"}
	listCmd := listIdpProviders(client)
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