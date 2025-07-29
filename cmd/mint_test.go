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

	// Check legacy flags are set up
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag)

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)

	idpFlag := cmd.Flags().Lookup("idp")
	assert.NotNil(t, idpFlag)

	keysetFlag := cmd.Flags().Lookup("keyset")
	assert.NotNil(t, keysetFlag)

	// Check new key-based flags are set up
	keysFlag := cmd.Flags().Lookup("keys")
	assert.NotNil(t, keysFlag)

	durationFlag := cmd.Flags().Lookup("duration")
	assert.NotNil(t, durationFlag)

	allFlag := cmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
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

	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", "", nil, 0, false)

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

	err := mintCredentialsWithFlags(client, cmd, "test-token", "json", "test-idp", "", nil, 0, false)

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
	_ = os.Unsetenv("OIDC_TOKEN")
	_ = os.Unsetenv("GITHUB_TOKEN")

	err := mintCredentialsWithFlags(client, cmd, "", "env", "", "", nil, 0, false)

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
	err := mintCredentialsWithFlags(client, cmd, "", "env", "hello-world", "", nil, 0, false)

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
			_ = os.Unsetenv("OIDC_TOKEN")
			_ = os.Unsetenv("GITHUB_TOKEN")

			// Set test environment variable
			_ = os.Setenv(tt.envVar, tt.envValue)
			defer func() { _ = os.Unsetenv(tt.envVar) }()

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

			err := mintCredentialsWithFlags(client, cmd, "", "env", "", "", nil, 0, false)

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

	outputAsEnvVars(creds, nil, cmd)

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

	outputAsEnvVars(creds, nil, cmd)

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

func TestMintCredentialsWithFlags_WithKeyset(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedCreds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2023-06-01T12:00:00Z",
	}

	credResponseBody, _ := json.Marshal(expectedCreds)
	credResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(credResponseBody)),
	}

	keysetKeys := map[string]string{
		"MINIO_ADMIN_ROLE": "minio:admin",
		"ACCESS_LEVEL":     "admin",
	}
	keysetResponseBody, _ := json.Marshal(keysetKeys)
	keysetResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(keysetResponseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(credResp, nil)
	mockClient.On("Get", "http://localhost:3000/credentials/keysets/keys?token=test-token&keyset=admin").Return(keysetResp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", "admin", nil, 0, false)

	assert.NoError(t, err)

	// Check standard AWS credentials in output
	assert.Contains(t, stdout.String(), "export AWS_ACCESS_KEY_ID=AKIATEST123")
	assert.Contains(t, stdout.String(), "export AWS_SECRET_ACCESS_KEY=secretkey123")

	// Check keyset environment variables in output
	assert.Contains(t, stdout.String(), "export MINIO_ADMIN_ROLE=minio:admin")
	assert.Contains(t, stdout.String(), "export ACCESS_LEVEL=admin")

	// Check keyset info in stderr
	assert.Contains(t, stderr.String(), "🔑 [LEGACY] Using keyset: admin")
	assert.Contains(t, stderr.String(), "🔑 Setting keyset environment variables")
	assert.Contains(t, stderr.String(), "MINIO_ADMIN_ROLE=minio:admin")
	assert.Contains(t, stderr.String(), "ACCESS_LEVEL=admin")
}

// New key-based tests

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

	mockClient.On("Post", "http://localhost:3000/credentials/mint-keys", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS", "AWS_CREDENTIALS"}
	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", "", keys, 0, false)

	assert.NoError(t, err)

	// Check output contains expected environment variables from both keys
	output := stdout.String()
	assert.Contains(t, output, "export MINIO_ACCESS_KEY_ID=AKIAMINIO123")
	assert.Contains(t, output, "export MINIO_SECRET_ACCESS_KEY=miniosecret123")
	assert.Contains(t, output, "export AWS_ACCESS_KEY_ID=AKIAAWS123")
	assert.Contains(t, output, "export AWS_SECRET_ACCESS_KEY=awssecret123")

	// Check stderr shows key-based approach info
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "🔑 Minting keys: [MINIO_CREDENTIALS AWS_CREDENTIALS]")
	assert.Contains(t, stderrOutput, "✅ Successfully minted 2 keys with 8 environment variables")

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

	mockClient.On("Post", "http://localhost:3000/credentials/mint-keys", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", "", nil, 0, true)

	assert.NoError(t, err)

	// Check stderr shows all flag info
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "🔑 Minting all available keys")
	assert.Contains(t, stderrOutput, "✅ Successfully minted 1 keys with 2 environment variables")

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

	mockClient.On("Post", "http://localhost:3000/credentials/mint-keys", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS"}
	duration := 1800 // 30 minutes
	err := mintCredentialsWithFlags(client, cmd, "test-token", "env", "test-idp", "", keys, duration, false)

	assert.NoError(t, err)

	// Check stderr shows duration override info
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "⏱️ Duration override: 1800 seconds")

	mockClient.AssertExpectations(t)
}

func TestMintCredentialsWithFlags_KeyBasedJSONOutput(t *testing.T) {
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

	mockClient.On("Post", "http://localhost:3000/credentials/mint-keys", "application/json", mock.Anything).Return(resp, nil)

	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	keys := []string{"MINIO_CREDENTIALS"}
	err := mintCredentialsWithFlags(client, cmd, "test-token", "json", "test-idp", "", keys, 0, false)

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

	// Check stderr for success message and key info
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "🔑 Key: MINIO_CREDENTIALS (expires: 2025-01-01T12:00:00Z)")
	assert.Contains(t, stderrOutput, "🔑 Key: AWS_CREDENTIALS (expires: 2025-01-01T12:00:00Z)")
	assert.Contains(t, stderrOutput, "✅ Successfully minted 2 keys with 5 environment variables")
	assert.Contains(t, stderrOutput, "💡 To use: eval \"$(voidkey mint --keys MINIO_CREDENTIALS)\"")
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
