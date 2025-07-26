package cmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient for mint tests
type MockHTTPClientMint struct {
	mock.Mock
}

func (m *MockHTTPClientMint) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientMint) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

// executeCommand helper function to capture output from Cobra commands
func executeCommand(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	
	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)
	
	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestMintCmd_HelloWorldIdP(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "hello-world IdP with default token",
			args: []string{"--idp", "hello-world"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "🎭 Using hello-world IdP with default token\n🔍 Using IdP provider: hello-world\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
		{
			name: "hello-world IdP with JSON output",
			args: []string{"--idp", "hello-world", "--output", "json"},
			expectedStdout: `{
  "accessKey": "AKIAIOSFODNN7EXAMPLE",
  "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "sessionToken": "hello-world-session-token",
  "expiresAt": "2025-07-25T18:00:00.000Z"
}
`,
			expectedStderr: "🎭 Using hello-world IdP with default token\n🔍 Using IdP provider: hello-world\n",
		},
		{
			name: "hello-world IdP with custom token",
			args: []string{"--idp", "hello-world", "--token", "custom-test-token"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "🔍 Using IdP provider: hello-world\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClientMint)
			
			// Setup expected response
			serverResponse := `{
				"accessKey": "AKIAIOSFODNN7EXAMPLE",
				"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"sessionToken": "hello-world-session-token",
				"expiresAt": "2025-07-25T18:00:00.000Z"
			}`
			
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(serverResponse)),
			}
			
			mockClient.On("Post", mock.AnythingOfType("string"), "application/json", mock.Anything).
				Return(resp, nil)

			// Create mint command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Execute command
			stdout, stderr, err := executeCommand(mintCmd, tt.args...)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStdout, stdout)
			assert.Equal(t, tt.expectedStderr, stderr)
			
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMintCmd_WithProvidedToken(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "explicit token with default IdP",
			args: []string{"--token", "test-token-123"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "🔍 Using server default IdP provider\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
		{
			name: "explicit token with specific IdP",
			args: []string{"--token", "test-token-123", "--idp", "auth0-test"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "🔍 Using IdP provider: auth0-test\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClientMint)
			
			// Setup expected response
			serverResponse := `{
				"accessKey": "AKIAIOSFODNN7EXAMPLE",
				"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"sessionToken": "hello-world-session-token",
				"expiresAt": "2025-07-25T18:00:00.000Z"
			}`
			
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(serverResponse)),
			}
			
			mockClient.On("Post", mock.AnythingOfType("string"), "application/json", mock.Anything).
				Return(resp, nil)

			// Create mint command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Execute command
			stdout, stderr, err := executeCommand(mintCmd, tt.args...)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStdout, stdout)
			assert.Equal(t, tt.expectedStderr, stderr)
			
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMintCmd_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		envVar         string
		envValue       string
		expectedStderr string
	}{
		{
			name:           "OIDC_TOKEN environment variable",
			args:           []string{},
			envVar:         "OIDC_TOKEN",
			envValue:       "env-oidc-token",
			expectedStderr: "🔍 Using OIDC_TOKEN environment variable\n🔍 Using server default IdP provider\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
		{
			name:           "GITHUB_TOKEN environment variable",
			args:           []string{},
			envVar:         "GITHUB_TOKEN",
			envValue:       "env-github-token",
			expectedStderr: "🔍 Using GITHUB_TOKEN environment variable\n🔍 Using server default IdP provider\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			oldValue := os.Getenv(tt.envVar)
			os.Setenv(tt.envVar, tt.envValue)
			defer func() {
				if oldValue == "" {
					os.Unsetenv(tt.envVar)
				} else {
					os.Setenv(tt.envVar, oldValue)
				}
			}()

			// Create mock client
			mockClient := new(MockHTTPClientMint)
			
			// Setup expected response
			serverResponse := `{
				"accessKey": "AKIAIOSFODNN7EXAMPLE",
				"secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"sessionToken": "hello-world-session-token",
				"expiresAt": "2025-07-25T18:00:00.000Z"
			}`
			
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(serverResponse)),
			}
			
			mockClient.On("Post", mock.AnythingOfType("string"), "application/json", mock.Anything).
				Return(resp, nil)

			// Create mint command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Execute command
			_, stderr, err := executeCommand(mintCmd, tt.args...)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStderr, stderr)
			
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMintCmd_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:          "no token provided and not hello-world",
			args:          []string{},
			expectError:   true,
			errorContains: "OIDC token is required",
		},
		{
			name:          "no token provided with non-hello-world IdP",
			args:          []string{"--idp", "auth0-test"},
			expectError:   true,
			errorContains: "OIDC token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client (won't be called due to early error)
			mockClient := new(MockHTTPClientMint)
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Execute command
			_, _, err := executeCommand(mintCmd, tt.args...)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMintCmd_ServerErrors(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		serverError    error
		serverResponse string
		statusCode     int
		expectError    bool
		errorContains  string
	}{
		{
			name:          "network error",
			args:          []string{"--token", "test-token"},
			serverError:   assert.AnError,
			expectError:   true,
			errorContains: "failed to connect to broker server",
		},
		{
			name:           "server error 500",
			args:           []string{"--token", "test-token"},
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "server returned error 500",
		},
		{
			name:           "invalid JSON response",
			args:           []string{"--token", "test-token"},
			serverResponse: "invalid json",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to parse credentials response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClientMint)

			if tt.serverError != nil {
				mockClient.On("Post", mock.AnythingOfType("string"), "application/json", mock.Anything).
					Return((*http.Response)(nil), tt.serverError)
			} else {
				resp := &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.serverResponse)),
				}
				mockClient.On("Post", mock.AnythingOfType("string"), "application/json", mock.Anything).
					Return(resp, nil)
			}

			// Create mint command with mock client
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Execute command
			_, _, err := executeCommand(mintCmd, tt.args...)

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

func TestMintCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedToken  string
		expectedFormat string
		expectedIdP    string
	}{
		{
			name:           "default values",
			args:           []string{},
			expectedToken:  "",
			expectedFormat: "env",
			expectedIdP:    "",
		},
		{
			name:           "custom token",
			args:           []string{"--token", "my-token"},
			expectedToken:  "my-token",
			expectedFormat: "env",
			expectedIdP:    "",
		},
		{
			name:           "json output format",
			args:           []string{"--output", "json"},
			expectedToken:  "",
			expectedFormat: "json",
			expectedIdP:    "",
		},
		{
			name:           "short flag for output",
			args:           []string{"-o", "json"},
			expectedToken:  "",
			expectedFormat: "json",
			expectedIdP:    "",
		},
		{
			name:           "specific IdP",
			args:           []string{"--idp", "hello-world"},
			expectedToken:  "",
			expectedFormat: "env",
			expectedIdP:    "hello-world",
		},
		{
			name:           "all flags combined",
			args:           []string{"--token", "test-token", "--output", "json", "--idp", "auth0-test"},
			expectedToken:  "test-token",
			expectedFormat: "json",
			expectedIdP:    "auth0-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client (won't be called due to parsing-only test)
			mockClient := new(MockHTTPClientMint)
			voidkeyClient := NewVoidkeyClient(mockClient, "http://test-server:3000")
			mintCmd := mintCreds(voidkeyClient)

			// Parse flags without executing
			mintCmd.SetArgs(tt.args)
			err := mintCmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			// Check flag values
			tokenFlag, err := mintCmd.Flags().GetString("token")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedToken, tokenFlag)

			outputFlag, err := mintCmd.Flags().GetString("output")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFormat, outputFlag)

			idpFlag, err := mintCmd.Flags().GetString("idp")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedIdP, idpFlag)
		})
	}
}