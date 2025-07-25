package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of HTTPClient
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestVoidkeyClient_MintCredentials(t *testing.T) {
	tests := []struct {
		name           string
		oidcToken      string
		serverResponse string
		statusCode     int
		serverError    error
		expected       *CloudCredentials
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful credential minting",
			oidcToken: "valid-token",
			serverResponse: `{
				"accessKey": "AKIAIOSFODNN7EXAMPLE",
				"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"sessionToken": "hello-world-session-token",
				"expiresAt": "2025-07-25T18:00:00.000Z"
			}`,
			statusCode: 200,
			expected: &CloudCredentials{
				AccessKey:    "AKIAIOSFODNN7EXAMPLE",
				SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				SessionToken: "hello-world-session-token",
				ExpiresAt:    "2025-07-25T18:00:00.000Z",
			},
			expectError: false,
		},
		{
			name:      "empty OIDC token",
			oidcToken: "",
			serverResponse: `{
				"accessKey": "AKIAIOSFODNN7EXAMPLE",
				"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"sessionToken": "hello-world-session-token",
				"expiresAt": "2025-07-25T18:00:00.000Z"
			}`,
			statusCode: 200,
			expected: &CloudCredentials{
				AccessKey:    "AKIAIOSFODNN7EXAMPLE",
				SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				SessionToken: "hello-world-session-token",
				ExpiresAt:    "2025-07-25T18:00:00.000Z",
			},
			expectError: false,
		},
		{
			name:           "server returns 400 error",
			oidcToken:      "invalid-token",
			serverResponse: "Bad Request: Invalid OIDC token",
			statusCode:     400,
			expectError:    true,
			errorContains:  "server returned error 400",
		},
		{
			name:           "server returns 500 error",
			oidcToken:      "valid-token",
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "server returned error 500",
		},
		{
			name:           "invalid JSON response",
			oidcToken:      "valid-token",
			serverResponse: "invalid json response",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to parse credentials response",
		},
		{
			name:          "network error",
			oidcToken:     "valid-token",
			serverError:   assert.AnError,
			expectError:   true,
			errorContains: "failed to connect to broker server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")

			if tt.serverError != nil {
				mockClient.On("Post", "http://test-server:3000/credentials/mint", "application/json", mock.Anything).
					Return((*http.Response)(nil), tt.serverError)
			} else {
				resp := &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.serverResponse)),
				}
				mockClient.On("Post", "http://test-server:3000/credentials/mint", "application/json", mock.Anything).
					Return(resp, nil)
			}

			result, err := voidkeyClient.MintCredentials(tt.oidcToken)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestVoidkeyClient_MintCredentials_RequestBody(t *testing.T) {
	mockClient := new(MockHTTPClient)
	voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")

	expectedResponse := `{
		"accessKey": "AKIAIOSFODNN7EXAMPLE",
		"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"sessionToken": "hello-world-session-token",
		"expiresAt": "2025-07-25T18:00:00.000Z"
	}`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(expectedResponse)),
	}

	// Use a custom matcher to verify the request body
	mockClient.On("Post", "http://test-server:3000/credentials/mint", "application/json", 
		mock.MatchedBy(func(body io.Reader) bool {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(body)
			if err != nil {
				return false
			}
			bodyStr := buf.String()
			return strings.Contains(bodyStr, "test-token") && strings.Contains(bodyStr, "oidcToken")
		})).Return(resp, nil)

	_, err := voidkeyClient.MintCredentials("test-token")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}