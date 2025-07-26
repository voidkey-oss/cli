package cmd

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient for list tests
type MockHTTPClientList struct {
	mock.Mock
}

func (m *MockHTTPClientList) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientList) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestListIdpProvidersCmd_SuccessfulExecution(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "multiple providers with default",
			serverResponse: `[
				{"name": "auth0-test", "isDefault": true},
				{"name": "hello-world", "isDefault": false},
				{"name": "okta-corporate", "isDefault": false}
			]`,
			expectedStdout: "NAME            DEFAULT\n----            -------\nauth0-test      ✓\nhello-world     \nokta-corporate  \n",
			expectedStderr: "",
		},
		{
			name:           "single provider",
			serverResponse: `[{"name": "hello-world", "isDefault": true}]`,
			expectedStdout: "NAME         DEFAULT\n----         -------\nhello-world  ✓\n",
			expectedStderr: "",
		},
		{
			name:           "no providers",
			serverResponse: `[]`,
			expectedStdout: "No Identity Providers configured\n",
			expectedStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClientList)
			
			// Setup expected response
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(tt.serverResponse)),
			}
			
			mockClient.On("Get", "http://test-server:3000/credentials/idp-providers").
				Return(resp, nil)

			// Create list command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			listCmd := listIdpProviders(voidkeyClient)

			// Execute command
			stdout, stderr, err := executeCommand(listCmd)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStdout, stdout)
			assert.Equal(t, tt.expectedStderr, stderr)
			
			mockClient.AssertExpectations(t)
		})
	}
}

func TestListIdpProvidersCmd_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverError    error
		serverResponse string
		statusCode     int
		expectError    bool
		errorContains  string
	}{
		{
			name:          "network error",
			serverError:   assert.AnError,
			expectError:   true,
			errorContains: "failed to list IdP providers",
		},
		{
			name:           "server error 500",
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "failed to list IdP providers",
		},
		{
			name:           "invalid JSON response",
			serverResponse: "invalid json",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to list IdP providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClientList)

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

			// Create list command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			listCmd := listIdpProviders(voidkeyClient)

			// Execute command
			_, _, err := executeCommand(listCmd)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}