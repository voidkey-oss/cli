package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHelloWorldIdP_Integration tests the hello-world IdP end-to-end functionality
func TestHelloWorldIdP_Integration(t *testing.T) {
	// Create a test server that mimics the voidkey broker-server with hello-world IdP
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/credentials/idp-providers":
			// Return hello-world provider in the list
			providers := []IdpProvider{
				{Name: "hello-world", IsDefault: true},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(providers)

		case "/credentials/mint":
			// Verify request method and headers
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Parse request body
			var req MintRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)

			// Verify hello-world behavior - it should accept any token
			assert.NotEmpty(t, req.OidcToken, "Token should be provided")

			// Return mock credentials (simulating hello-world IdP response)
			credentials := CloudCredentials{
				AccessKey:    "AKIAIOSFODNN7EXAMPLE",
				SecretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				SessionToken: "hello-world-session-token",
				ExpiresAt:    "2025-07-25T20:00:00.000Z",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(credentials)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create client pointing to test server
	voidkeyClient := NewVoidkeyClient(server.Client(), server.URL)

	t.Run("list providers shows hello-world", func(t *testing.T) {
		listCmd := listIdpProviders(voidkeyClient)
		stdout, stderr, err := executeCommand(listCmd)

		assert.NoError(t, err)
		assert.Empty(t, stderr)
		assert.Contains(t, stdout, "hello-world")
		assert.Contains(t, stdout, "✓") // Should show as default
	})

	t.Run("mint with hello-world IdP", func(t *testing.T) {
		mintCmd := mintCreds(voidkeyClient)
		stdout, stderr, err := executeCommand(mintCmd, "--idp", "hello-world")

		assert.NoError(t, err)
		assert.Contains(t, stderr, "Using hello-world IdP with default token")
		assert.Contains(t, stdout, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		assert.Contains(t, stdout, "AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		assert.Contains(t, stdout, "AWS_SESSION_TOKEN=hello-world-session-token")
	})

	t.Run("mint with hello-world IdP and custom token", func(t *testing.T) {
		mintCmd := mintCreds(voidkeyClient)
		stdout, stderr, err := executeCommand(mintCmd, "--idp", "hello-world", "--token", "my-custom-test-token")

		assert.NoError(t, err)
		assert.Contains(t, stderr, "Using IdP provider: hello-world")
		assert.Contains(t, stdout, "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")
		assert.NotContains(t, stderr, "Using hello-world IdP with default token") // Should not use default token
	})

	t.Run("mint with hello-world IdP json output", func(t *testing.T) {
		mintCmd := mintCreds(voidkeyClient)
		stdout, stderr, err := executeCommand(mintCmd, "--idp", "hello-world", "--output", "json")

		assert.NoError(t, err)
		assert.Contains(t, stderr, "Using hello-world IdP with default token")
		
		// Verify JSON structure
		var credentials CloudCredentials
		err = json.Unmarshal([]byte(stdout), &credentials)
		assert.NoError(t, err)
		assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", credentials.AccessKey)
		assert.Equal(t, "hello-world-session-token", credentials.SessionToken)
	})
}

// TestHelloWorldIdP_ClientMethods tests the client methods directly
func TestHelloWorldIdP_ClientMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/credentials/idp-providers":
			providers := []IdpProvider{
				{Name: "hello-world", IsDefault: true},
				{Name: "auth0-test", IsDefault: false},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(providers)
		}
	}))
	defer server.Close()

	client := NewVoidkeyClient(server.Client(), server.URL)

	t.Run("list providers includes hello-world", func(t *testing.T) {
		providers, err := client.ListIdpProviders()
		
		assert.NoError(t, err)
		assert.Len(t, providers, 2)
		
		// Find hello-world provider
		var helloWorldProvider *IdpProvider
		for _, p := range providers {
			if p.Name == "hello-world" {
				helloWorldProvider = &p
				break
			}
		}
		
		assert.NotNil(t, helloWorldProvider, "hello-world provider should be in the list")
		assert.True(t, helloWorldProvider.IsDefault, "hello-world should be the default provider")
	})
}