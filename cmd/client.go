package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient interface for dependency injection and testing
type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// VoidkeyClient handles communication with the Voidkey broker server
type VoidkeyClient struct {
	client    HTTPClient
	serverURL string
}

// NewVoidkeyClient creates a new client with the given HTTP client and server URL
func NewVoidkeyClient(client HTTPClient, serverURL string) *VoidkeyClient {
	return &VoidkeyClient{
		client:    client,
		serverURL: serverURL,
	}
}

// MintCredentials calls the broker server to mint credentials
func (c *VoidkeyClient) MintCredentials(oidcToken string, idpName string) (*CloudCredentials, error) {
	// Prepare request
	reqBody := MintRequest{
		OidcToken: oidcToken,
		IdpName:   idpName,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/mint", c.serverURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var credentials CloudCredentials
	if err := json.Unmarshal(body, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credentials response: %w", err)
	}

	return &credentials, nil
}

// IdpProvider represents an identity provider
type IdpProvider struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// ListIdpProviders calls the broker server to list available IdP providers
func (c *VoidkeyClient) ListIdpProviders() ([]IdpProvider, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/idp-providers", c.serverURL)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var providers []IdpProvider
	if err := json.Unmarshal(body, &providers); err != nil {
		return nil, fmt.Errorf("failed to parse providers response: %w", err)
	}

	return providers, nil
}