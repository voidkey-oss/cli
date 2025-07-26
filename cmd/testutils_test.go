package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

// TestHelpers provides common testing utilities for CLI tests

// CreateMockHTTPResponse creates a mock HTTP response for testing
func CreateMockHTTPResponse(statusCode int, body interface{}) *http.Response {
	var bodyReader io.Reader
	
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			// Assume it's a struct that should be JSON marshalled
			jsonData, _ := json.Marshal(v)
			bodyReader = bytes.NewReader(jsonData)
		}
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bodyReader),
	}
}

// SetupTestCommand creates a test command with captured output
func SetupTestCommand() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}

// CreateTestCredentials returns sample credentials for testing
func CreateTestCredentials() CloudCredentials {
	return CloudCredentials{
		AccessKey:    "AKIATEST123456789",
		SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		SessionToken: "AQoEXAMPLEH4aoAH0gNCAPyJxz4BlCFFxWNE1OPTgk5TthT+FvwqnKwRcOIfrRh3c/LTo6UDdyJwOOvEVPvLXCrrrUtdnniCEXAMPLE/IvU1dYUg2RVAJBanLiHb4IgRmpRV3zrkuWJOgQs8IZZaIv2BXIa2R4OlgkBN9bkUDNCJiBeb/AXlzBBko7b15fjrBs2+cTQtpZ3CYWFXG8C5zqx37wnOE49mRl/+OtkIKGO7fAE",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}
}

// CreateTestIdpProviders returns sample IdP providers for testing
func CreateTestIdpProviders() []IdpProvider {
	return []IdpProvider{
		{Name: "auth0", IsDefault: true},
		{Name: "github", IsDefault: false},
		{Name: "okta", IsDefault: false},
		{Name: "hello-world", IsDefault: false},
	}
}

// AssertCommandOutput checks common command output patterns
func AssertCommandOutput(t *testing.T, stdout, stderr *bytes.Buffer, expectedStdout, expectedStderr []string) {
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	
	for _, expected := range expectedStdout {
		if expected != "" && !bytes.Contains(stdout.Bytes(), []byte(expected)) {
			t.Errorf("Expected stdout to contain '%s', got: %s", expected, stdoutStr)
		}
	}
	
	for _, expected := range expectedStderr {
		if expected != "" && !bytes.Contains(stderr.Bytes(), []byte(expected)) {
			t.Errorf("Expected stderr to contain '%s', got: %s", expected, stderrStr)
		}
	}
}

// MockSuccessfulMintResponse sets up a mock for successful credential minting
func MockSuccessfulMintResponse(mockClient *MockHTTPClient, serverURL string, credentials CloudCredentials) {
	resp := CreateMockHTTPResponse(http.StatusOK, credentials)
	mockClient.On("Post", serverURL+"/credentials/mint", "application/json", mock.Anything).Return(resp, nil)
}

// MockSuccessfulListResponse sets up a mock for successful provider listing
func MockSuccessfulListResponse(mockClient *MockHTTPClient, serverURL string, providers []IdpProvider) {
	resp := CreateMockHTTPResponse(http.StatusOK, providers)
	mockClient.On("Get", serverURL+"/credentials/idp-providers").Return(resp, nil)
}

// MockErrorResponse sets up a mock for error responses
func MockErrorResponse(mockClient *MockHTTPClient, method, url string, statusCode int, errorMessage string) {
	resp := CreateMockHTTPResponse(statusCode, errorMessage)
	
	switch method {
	case "POST":
		mockClient.On("Post", url, "application/json", mock.Anything).Return(resp, nil)
	case "GET":
		mockClient.On("Get", url).Return(resp, nil)
	}
}

// TestMockHTTPClient_Implementation verifies the mock implementation works correctly
func TestMockHTTPClient_Implementation(t *testing.T) {
	mockClient := &MockHTTPClient{}
	
	// Test POST method
	resp := CreateMockHTTPResponse(200, "test body")
	mockClient.On("Post", "http://test.com", "application/json", mock.Anything).Return(resp, nil)
	
	result, err := mockClient.Post("http://test.com", "application/json", bytes.NewReader([]byte("test")))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", result.StatusCode)
	}
	
	mockClient.AssertExpectations(t)
}

// TestCreateMockHTTPResponse verifies the response creation utility
func TestCreateMockHTTPResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       interface{}
		expected   string
	}{
		{
			name:       "string body",
			statusCode: 200,
			body:       "test string",
			expected:   "test string",
		},
		{
			name:       "struct body",
			statusCode: 200,
			body:       map[string]string{"key": "value"},
			expected:   `{"key":"value"}`,
		},
		{
			name:       "nil body",
			statusCode: 404,
			body:       nil,
			expected:   "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := CreateMockHTTPResponse(tt.statusCode, tt.body)
			
			if resp.StatusCode != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, resp.StatusCode)
			}
			
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Error reading body: %v", err)
			}
			
			if string(body) != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, string(body))
			}
		})
	}
}

// TestCreateTestCredentials verifies the test credentials helper
func TestCreateTestCredentials(t *testing.T) {
	creds := CreateTestCredentials()
	
	if creds.AccessKey == "" {
		t.Error("AccessKey should not be empty")
	}
	if creds.SecretKey == "" {
		t.Error("SecretKey should not be empty")
	}
	if creds.ExpiresAt == "" {
		t.Error("ExpiresAt should not be empty")
	}
}

// TestCreateTestIdpProviders verifies the test providers helper
func TestCreateTestIdpProviders(t *testing.T) {
	providers := CreateTestIdpProviders()
	
	if len(providers) == 0 {
		t.Error("Should return at least one provider")
	}
	
	// Check that exactly one provider is marked as default
	defaultCount := 0
	for _, provider := range providers {
		if provider.IsDefault {
			defaultCount++
		}
	}
	
	if defaultCount != 1 {
		t.Errorf("Expected exactly 1 default provider, got %d", defaultCount)
	}
}