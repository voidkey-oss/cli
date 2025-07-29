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

	// Check key-based flags are set up
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag)

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)

	idpFlag := cmd.Flags().Lookup("idp")
	assert.NotNil(t, idpFlag)

	keysFlag := cmd.Flags().Lookup("keys")
	assert.NotNil(t, keysFlag)

	durationFlag := cmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)

	allFlag := cmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
}

func TestMintCredentialsWithFlags_KeyBasedApproach(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID":     "AKIAMINIO123",
				"MINIO_SECRET_ACCESS_KEY": "miniosecret123",
				"MINIO_SESSION_TOKEN":     "miniosession123",
				"MINIO_EXPIRATION":        "2025-01-01T12:00:00Z",
				"MINIO_ENDPOINT":          "http://localhost:9000",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
			Metadata:  map[string]any{"provider": "minio-test", "keyName": "MINIO_CREDENTIALS"},
		},
		"AWS_CREDENTIALS": {
			Credentials: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAAWS123",
				"AWS_SECRET_ACCESS_KEY": "awssecret123",
				"AWS_SESSION_TOKEN":     "awssession123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
			Metadata:  map[string]any{"provider": "aws-test", "keyName": "AWS_CREDENTIALS"},
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS", "AWS_CREDENTIALS"}
	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", keys, 0, false)

	assert.NoError(t, err)

	// Check output contains expected environment variables from both keys
	output := stdout.String()
	assert.Contains(t, output, "export MINIO_ACCESS_KEY_ID=AKIAMINIO123")
	assert.Contains(t, output, "export MINIO_SECRET_ACCESS_KEY=miniosecret123")
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIAAWS123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=awssecret123")

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_AllFlag(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID": "AKIAMINIO123",
				"MINIO_ENDPOINT":      "http://localhost:9000",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", nil, 0, true)

	assert.NoError(t, err)

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
	_ = os.Unsetenv("OIDC_TOKEN")
	_ = os.Unsetenv("GITHUB_TOKEN")

	err := mintCredentialsWithFlags(client, cmd, "", "env", "", nil, 0, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OIDC token is required")
}

func TestMintCredentialsWithFlags_JSONOutput(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID": "AKIAMINIO123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS"}
	err := mintCredentialsWithFlags(client, cmd, "test-token", "json", "test-idp", keys, 0, false)

	assert.NoError(t, err)

	// Check output is valid JSON
	output := stdout.String()
	var parsedResponse map[string]KeyCredentialResponse
	err = json.Unmarshal([]byte(output), &parsedResponse)
	assert.NoError(t, err)
	assert.Contains(t, parsedResponse, "MINIO_CREDENTIALS")
	assert.Equal(t, "AKIAMINIO123", parsedResponse["MINIO_CREDENTIALS"].Credentials["MINIO_ACCESS_KEY_ID"])

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_WithDuration(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID": "AKIAMINIO123",
			},
			ExpiresAt: "2025-01-01T12:30:00Z", // 30 minutes from minting
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS"}
	duration := 1800 // 30 minutes
	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", keys, duration, false)

	assert.NoError(t, err)

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
			_ = os.Unsetenv("OIDC_TOKEN")
			_ = os.Unsetenv("GITHUB_TOKEN")

			// Set test environment variable
			_ = os.Setenv(tt.envVar, tt.envValue)
			defer func() { _ = os.Unsetenv(tt.envVar) }()

			mockClient := &MockHTTPClient{}
			client := NewVoidkeyClient(mockClient, "http://localhost:3000")

			expectedKeyResponses := map[string]KeyCredentialResponse{
				"MINIO_CREDENTIALS": {
					Credentials: map[string]string{
						"MINIO_ACCESS_KEY_ID": "AKIATEST123",
					},
					ExpiresAt: "2024-12-31T23:59:59Z",
				},
			}

			responseBody, _ := json.Marshal(expectedKeyResponses)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}

			mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

			var stdout, stderr bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := mintCredentialsWithFlags(client, cmd, "", "env", "", []string{"MINIO_CREDENTIALS"}, 0, false)

			assert.NoError(t, err)

			// Check stderr for environment variable message
			stderrOutput := stderr.String()
			assert.Contains(t, stderrOutput, tt.expected)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestOutputKeysAsEnvVars(t *testing.T) {
	keyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID":     "AKIAMINIO123",
				"MINIO_SECRET_ACCESS_KEY": "miniosecret123",
				"MINIO_ENDPOINT":          "http://localhost:9000",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
		"AWS_CREDENTIALS": {
			Credentials: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAAWS123",
				"AWS_SECRET_ACCESS_KEY": "awssecret123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
	}

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	outputKeysAsEnvVars(keyResponses, cmd)

	output := stdout.String()
	assert.Contains(t, output, "export MINIO_ACCESS_KEY_ID=AKIAMINIO123")
	assert.Contains(t, output, "export MINIO_SECRET_ACCESS_KEY=miniosecret123")
	assert.Contains(t, output, "export MINIO_ENDPOINT=http://localhost:9000")
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIAAWS123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=awssecret123")
}

func TestOutputKeysAsJSON(t *testing.T) {
	keyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID": "AKIAMINIO123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
	}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	outputKeysAsJSON(keyResponses, cmd)

	output := stdout.String()

	// Parse JSON to verify it's valid
	var parsedResponse map[string]KeyCredentialResponse
	err := json.Unmarshal([]byte(output), &parsedResponse)
	assert.NoError(t, err)
	assert.Contains(t, parsedResponse, "MINIO_CREDENTIALS")
	assert.Equal(t, "AKIAMINIO123", parsedResponse["MINIO_CREDENTIALS"].Credentials["MINIO_ACCESS_KEY_ID"])

	// Check that output is properly formatted JSON
	assert.True(t, strings.Contains(output, "{\n"))
	assert.True(t, strings.Contains(output, "  \"MINIO_CREDENTIALS\":"))
}