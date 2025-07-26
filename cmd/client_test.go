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

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestVoidkeyClient_MintCredentials(t *testing.T) {
	tests := []struct {
		name           string
		oidcToken      string
		idpName        string
		serverResponse string
		statusCode     int
		serverError    error
		expected       *CloudCredentials
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful credential minting with token and IdP",
			oidcToken: "valid-token",
			idpName:   "auth0-test",
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
			name:      "successful credential minting with default IdP",
			oidcToken: "valid-token",
			idpName:   "",
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
			idpName:        "",
			serverResponse: "Bad Request: Invalid OIDC token",
			statusCode:     400,
			expectError:    true,
			errorContains:  "server returned error 400",
		},
		{
			name:           "server returns 500 error",
			oidcToken:      "valid-token",
			idpName:        "",
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "server returned error 500",
		},
		{
			name:           "invalid JSON response",
			oidcToken:      "valid-token",
			idpName:        "",
			serverResponse: "invalid json response",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to parse credentials response",
		},
		{
			name:          "network error",
			oidcToken:     "valid-token",
			idpName:       "",
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

			result, err := voidkeyClient.MintCredentials(tt.oidcToken, tt.idpName)

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
	tests := []struct {
		name          string
		oidcToken     string
		idpName       string
		expectedToken string
		expectedIdP   string
	}{
		{
			name:          "request with token and IdP",
			oidcToken:     "test-token-123",
			idpName:       "auth0-test",
			expectedToken: "test-token-123",
			expectedIdP:   "auth0-test",
		},
		{
			name:          "request with token only",
			oidcToken:     "test-token-456",
			idpName:       "",
			expectedToken: "test-token-456",
			expectedIdP:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
					
					// Check for expected token and IdP in request
					hasToken := strings.Contains(bodyStr, tt.expectedToken) && strings.Contains(bodyStr, "oidcToken")
					if tt.expectedIdP != "" {
						return hasToken && strings.Contains(bodyStr, tt.expectedIdP) && strings.Contains(bodyStr, "idpName")
					}
					return hasToken
				})).Return(resp, nil)

			_, err := voidkeyClient.MintCredentials(tt.oidcToken, tt.idpName)

			assert.NoError(t, err)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestVoidkeyClient_ListIdpProviders(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		serverError    error
		expected       []IdpProvider
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful provider listing",
			serverResponse: `[
				{"name": "hello-world", "isDefault": true},
				{"name": "okta-corporate", "isDefault": false},
				{"name": "auth0-test", "isDefault": false}
			]`,
			statusCode: 200,
			expected: []IdpProvider{
				{Name: "hello-world", IsDefault: true},
				{Name: "okta-corporate", IsDefault: false},
				{Name: "auth0-test", IsDefault: false},
			},
			expectError: false,
		},
		{
			name:           "empty provider list",
			serverResponse: `[]`,
			statusCode:     200,
			expected:       []IdpProvider{},
			expectError:    false,
		},
		{
			name:           "server returns 500 error",
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "server returned error 500",
		},
		{
			name:           "invalid JSON response",
			serverResponse: "invalid json response",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to parse providers response",
		},
		{
			name:          "network error",
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
				mockClient.On("Get", "http://test-server:3000/credentials/idp-providers").
					Return((*http.Response)(nil), tt.serverError)
			} else {
				resp := &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.serverResponse)),
				}
				mockClient.On("Get", "http://test-server:3000/credentials/idp-providers").
					Return(resp, nil)
			}

			result, err := voidkeyClient.ListIdpProviders()

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