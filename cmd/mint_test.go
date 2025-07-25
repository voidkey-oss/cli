package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestMintCmd_SuccessfulExecution(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "default env output with dummy token",
			args: []string{},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "🔄 Using dummy OIDC token for hello world demo\n✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
		{
			name: "json output format",
			args: []string{"--output", "json"},
			expectedStdout: `{
  "accessKey": "AKIAIOSFODNN7EXAMPLE",
  "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "sessionToken": "hello-world-session-token",
  "expiresAt": "2025-07-25T18:00:00.000Z"
}
`,
			expectedStderr: "🔄 Using dummy OIDC token for hello world demo\n",
		},
		{
			name: "custom token provided",
			args: []string{"--token", "custom-token"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=hello-world-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T18:00:00.000Z
`,
			expectedStderr: "✅ Credentials minted successfully (expires: 2025-07-25T18:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClient)
			
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

func TestMintCmd_ErrorHandling(t *testing.T) {
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
			args:          []string{},
			serverError:   assert.AnError,
			expectError:   true,
			errorContains: "failed to connect to broker server",
		},
		{
			name:           "server error 500",
			args:           []string{},
			serverResponse: "Internal Server Error",
			statusCode:     500,
			expectError:    true,
			errorContains:  "server returned error 500",
		},
		{
			name:           "invalid JSON response",
			args:           []string{},
			serverResponse: "invalid json",
			statusCode:     200,
			expectError:    true,
			errorContains:  "failed to parse credentials response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockHTTPClient)

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
		name          string
		args          []string
		expectedToken string
		expectedFormat string
	}{
		{
			name:           "default values",
			args:           []string{},
			expectedToken:  "",
			expectedFormat: "env",
		},
		{
			name:           "custom token",
			args:           []string{"--token", "my-token"},
			expectedToken:  "my-token",
			expectedFormat: "env",
		},
		{
			name:           "json output format",
			args:           []string{"--output", "json"},
			expectedToken:  "",
			expectedFormat: "json",
		},
		{
			name:           "short flag for output",
			args:           []string{"-o", "json"},
			expectedToken:  "",
			expectedFormat: "json",
		},
		{
			name:           "both flags",
			args:           []string{"--token", "test-token", "--output", "json"},
			expectedToken:  "test-token",
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client (won't be called due to parsing-only test)
			mockClient := new(MockHTTPClient)
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
		})
	}
}