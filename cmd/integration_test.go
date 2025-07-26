package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMintCmd_Integration(t *testing.T) {
	// Create a test server that mimics the voidkey broker-server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/credentials/mint", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var req MintRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		// Return mock credentials
		credentials := CloudCredentials{
			AccessKey:    "AKIAIOSFODNN7EXAMPLE",
			SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			SessionToken: "integration-test-session-token",
			ExpiresAt:    "2025-07-25T20:00:00.000Z",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(credentials)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Create client pointing to test server
	voidkeyClient := NewVoidkeyClient(server.Client(), server.URL)

	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectError    bool
	}{
		{
			name: "successful env output with explicit token",
			args: []string{"--token", "integration-test-token"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=integration-test-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T20:00:00.000Z
`,
			expectedStderr: "🔍 Using server default IdP provider\n✅ Credentials minted successfully (expires: 2025-07-25T20:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
			expectError:    false,
		},
		{
			name: "successful json output with explicit token",
			args: []string{"--token", "integration-test-token", "--output", "json"},
			expectedStdout: `{
  "accessKey": "AKIAIOSFODNN7EXAMPLE",
  "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "sessionToken": "integration-test-session-token",
  "expiresAt": "2025-07-25T20:00:00.000Z"
}
`,
			expectedStderr: "🔍 Using server default IdP provider\n",
			expectError:    false,
		},
		{
			name: "successful with hello-world IdP (no token required)",
			args: []string{"--idp", "hello-world"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=integration-test-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T20:00:00.000Z
`,
			expectedStderr: "🎭 Using hello-world IdP with default token\n🔍 Using IdP provider: hello-world\n✅ Credentials minted successfully (expires: 2025-07-25T20:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
			expectError:    false,
		},
		{
			name: "successful with specific IdP and token",
			args: []string{"--token", "integration-test-token", "--idp", "auth0-test"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=integration-test-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T20:00:00.000Z
`,
			expectedStderr: "🔍 Using IdP provider: auth0-test\n✅ Credentials minted successfully (expires: 2025-07-25T20:00:00.000Z)\n💡 To use: eval \"$(voidkey mint)\"\n",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh command for each test to avoid flag persistence
			mintCmd := mintCreds(voidkeyClient)
			stdout, stderr, err := executeCommand(mintCmd, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStdout, stdout)
				assert.Equal(t, tt.expectedStderr, stderr)
			}
		})
	}
}

func TestMintCmd_IntegrationErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		serverFunc    http.HandlerFunc
		expectError   bool
		errorContains string
	}{
		{
			name: "server returns 400 error",
			args: []string{"--token", "invalid-token"},
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("Bad Request: Invalid OIDC token"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError:   true,
			errorContains: "server returned error 400",
		},
		{
			name: "server returns invalid JSON",
			args: []string{"--token", "valid-token"},
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("invalid json response"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError:   true,
			errorContains: "failed to parse credentials response",
		},
		{
			name: "server returns 500 error",
			args: []string{"--token", "valid-token"},
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Internal Server Error"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError:   true,
			errorContains: "server returned error 500",
		},
		{
			name:          "no token provided and not hello-world",
			args:          []string{},
			serverFunc:    nil, // Won't be called due to early error
			expectError:   true,
			errorContains: "OIDC token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverFunc != nil {
				server = httptest.NewServer(tt.serverFunc)
				defer server.Close()
			}

			var serverURL string
			if server != nil {
				serverURL = server.URL
			} else {
				serverURL = "http://test-server:3000" // Won't be used
			}

			voidkeyClient := NewVoidkeyClient(&http.Client{}, serverURL)
			mintCmd := mintCreds(voidkeyClient)

			_, _, err := executeCommand(mintCmd, tt.args...)

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

func TestListIdpProvidersCmd_Integration(t *testing.T) {
	// Create a test server that mimics the voidkey broker-server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/credentials/idp-providers", r.URL.Path)

		// Return mock providers
		providers := []IdpProvider{
			{Name: "hello-world", IsDefault: true},
			{Name: "okta-corporate", IsDefault: false},
			{Name: "auth0-test", IsDefault: false},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(providers)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Create client pointing to test server
	voidkeyClient := NewVoidkeyClient(server.Client(), server.URL)
	listCmd := listIdpProviders(voidkeyClient)

	// Execute command
	stdout, stderr, err := executeCommand(listCmd)

	// Assertions
	assert.NoError(t, err)
	assert.Empty(t, stderr)
	
	// Check that all providers are listed
	assert.Contains(t, stdout, "hello-world")
	assert.Contains(t, stdout, "okta-corporate")
	assert.Contains(t, stdout, "auth0-test")
	assert.Contains(t, stdout, "✓") // Default indicator
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "DEFAULT")
}

func TestListIdpProvidersCmd_IntegrationErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		serverFunc    http.HandlerFunc
		expectError   bool
		errorContains string
	}{
		{
			name: "server returns 500 error",
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Internal Server Error"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError:   true,
			errorContains: "failed to list IdP providers",
		},
		{
			name: "server returns invalid JSON",
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("invalid json response"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError:   true,
			errorContains: "failed to list IdP providers",
		},
		{
			name: "empty provider list",
			serverFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("[]"))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}),
			expectError: false, // Should handle empty list gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverFunc)
			defer server.Close()

			voidkeyClient := NewVoidkeyClient(server.Client(), server.URL)
			listCmd := listIdpProviders(voidkeyClient)

			stdout, _, err := executeCommand(listCmd)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.name == "empty provider list" {
					assert.Contains(t, stdout, "No Identity Providers configured")
				}
			}
		})
	}
}