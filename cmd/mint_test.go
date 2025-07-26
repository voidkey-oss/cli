package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMintCreds_CommandCreation(t *testing.T) {
	mockClient := &MockHTTPClient{}
	voidkeyClient := NewVoidkeyClient(mockClient, "http://localhost:3000")

	cmd := mintCreds(voidkeyClient)

	assert.NotNil(t, cmd)
	assert.Equal(t, "mint", cmd.Use)
	assert.Contains(t, cmd.Short, "Mint short-lived cloud credentials")
	
	// Check flags are set up
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag)
	
	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)
	
	idpFlag := cmd.Flags().Lookup("idp")
	assert.NotNil(t, idpFlag)
}

func TestMintCredentialsWithFlags_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedCreds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	responseBody, _ := json.Marshal(expectedCreds)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp")

	assert.NoError(t, err)
	
	// Check output contains expected environment variables
	output := stdout.String()
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIATEST123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=secretkey123")
	assert.Contains(t, output, "export AWS_SESSION_TOKEN=sessiontoken123")
	assert.Contains(t, output, "export AWS_CREDENTIAL_EXPIRATION=2024-12-31T23:59:59Z")

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_JSONOutput(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedCreds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	responseBody, _ := json.Marshal(expectedCreds)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := mintCredentialsWithFlags(client, cmd, "test-token", "json", "test-idp")

	assert.NoError(t, err)
	
	// Check output is valid JSON
	output := stdout.String()
	var parsedCreds CloudCredentials
	err = json.Unmarshal([]byte(output), &parsedCreds)
	assert.NoError(t, err)
	assert.Equal(t, expectedCreds.AccessKey, parsedCreds.AccessKey)

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_NoToken(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Clear any environment variables
	os.Unsetenv("OIDC_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	err := mintCredentialsWithFlags(client, cmd, "", "env", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OIDC token is required")
}

func TestMintCredentialsWithFlags_HelloWorldIdP(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedCreds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	responseBody, _ := json.Marshal(expectedCreds)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Test hello-world IdP with empty token (should use default)
	err := mintCredentialsWithFlags(client, cmd, "", "env", "hello-world")

	assert.NoError(t, err)
	
	// Check stderr for hello-world message
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "Using hello-world IdP with default token")

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_EnvironmentTokens(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "OIDC_TOKEN environment variable",
			envVar:   "OIDC_TOKEN",
			envValue: "oidc-test-token",
			expected: "Using OIDC_TOKEN environment variable",
		},
		{
			name:     "GITHUB_TOKEN environment variable",
			envVar:   "GITHUB_TOKEN",
			envValue: "github-test-token",
			expected: "Using GITHUB_TOKEN environment variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("OIDC_TOKEN")
			os.Unsetenv("GITHUB_TOKEN")

			// Set test environment variable
			os.Setenv(tt.envVar, tt.envValue)
			defer os.Unsetenv(tt.envVar)

			mockClient := &MockHTTPClient{}
			client := NewVoidkeyClient(mockClient, "http://localhost:3000")

			expectedCreds := CloudCredentials{
				AccessKey: "AKIATEST123",
				SecretKey: "secretkey123",
				ExpiresAt: "2024-12-31T23:59:59Z",
			}

			responseBody, _ := json.Marshal(expectedCreds)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}

			mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

			var stdout, stderr bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := mintCredentialsWithFlags(client, cmd, "", "env", "")

			assert.NoError(t, err)
			
			// Check stderr for environment variable message
			stderrOutput := stderr.String()
			assert.Contains(t, stderrOutput, tt.expected)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestOutputAsEnvVars(t *testing.T) {
	creds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	outputAsEnvVars(creds, cmd)

	output := stdout.String()
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIATEST123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=secretkey123")
	assert.Contains(t, output, "export AWS_SESSION_TOKEN=sessiontoken123")
	assert.Contains(t, output, "export AWS_CREDENTIAL_EXPIRATION=2024-12-31T23:59:59Z")

	// Check stderr for success message
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "Credentials minted successfully")
	assert.Contains(t, stderrOutput, "eval \"$(voidkey mint)\"")
}

func TestOutputAsEnvVars_NoSessionToken(t *testing.T) {
	creds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "", // No session token
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	outputAsEnvVars(creds, cmd)

	output := stdout.String()
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIATEST123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=secretkey123")
	assert.NotContains(t, output, "export AWS_SESSION_TOKEN=")
	assert.Contains(t, output, "export AWS_CREDENTIAL_EXPIRATION=2024-12-31T23:59:59Z")
}

func TestOutputAsJSON(t *testing.T) {
	creds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	outputAsJSON(creds, cmd)

	output := stdout.String()
	
	// Parse JSON to verify it's valid
	var parsedCreds CloudCredentials
	err := json.Unmarshal([]byte(output), &parsedCreds)
	assert.NoError(t, err)
	assert.Equal(t, creds.AccessKey, parsedCreds.AccessKey)
	assert.Equal(t, creds.SecretKey, parsedCreds.SecretKey)
	assert.Equal(t, creds.SessionToken, parsedCreds.SessionToken)
	assert.Equal(t, creds.ExpiresAt, parsedCreds.ExpiresAt)

	// Check that output is properly formatted JSON
	assert.True(t, strings.Contains(output, "{\n"))
	assert.True(t, strings.Contains(output, "  \"accessKey\":"))
}