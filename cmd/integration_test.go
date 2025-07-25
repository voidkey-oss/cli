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
	mintCmd := mintCreds(voidkeyClient)

	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectError    bool
	}{
		{
			name: "successful env output",
			args: []string{"--token", "integration-test-token"},
			expectedStdout: `export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=integration-test-session-token
export AWS_CREDENTIAL_EXPIRATION=2025-07-25T20:00:00.000Z
`,
			expectError: false,
		},
		{
			name: "successful json output",
			args: []string{"--token", "integration-test-token", "--output", "json"},
			expectedStdout: `{
  "accessKey": "AKIAIOSFODNN7EXAMPLE",
  "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "sessionToken": "integration-test-session-token",
  "expiresAt": "2025-07-25T20:00:00.000Z"
}
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := executeCommand(mintCmd, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStdout, stdout)
			}
		})
	}
}

func TestMintCmd_IntegrationErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		serverFunc  http.HandlerFunc
		expectError bool
		errorContains string
	}{
		{
			name: "server returns 400 error",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverFunc)
			defer server.Close()

			voidkeyClient := NewVoidkeyClient(server.Client(), server.URL)
			mintCmd := mintCreds(voidkeyClient)

			_, _, err := executeCommand(mintCmd, "--token", "test-token")

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