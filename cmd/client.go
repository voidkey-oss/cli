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
func (c *VoidkeyClient) MintCredentials(oidcToken string, idpName string, keyset string) (*CloudCredentials, error) {
	// Prepare request
	reqBody := MintRequest{
		OidcToken: oidcToken,
		IdpName:   idpName,
		Keyset:    keyset,
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
	defer func() {
		_ = resp.Body.Close()
	}()

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
	defer func() {
		_ = resp.Body.Close()
	}()

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

// GetAvailableKeysets calls the broker server to get available keysets for a subject
func (c *VoidkeyClient) GetAvailableKeysets(subject string) (map[string]map[string]string, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/keysets?subject=%s", c.serverURL, subject)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keysets map[string]map[string]string
	if err := json.Unmarshal(body, &keysets); err != nil {
		return nil, fmt.Errorf("failed to parse keysets response: %w", err)
	}

	return keysets, nil
}

// GetKeysetKeys calls the broker server to get keys for a specific keyset
func (c *VoidkeyClient) GetKeysetKeys(subject string, keysetName string) (map[string]string, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/keysets/keys?subject=%s&keyset=%s", c.serverURL, subject, keysetName)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keys map[string]string
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, fmt.Errorf("failed to parse keys response: %w", err)
	}

	return keys, nil
}

// GetAvailableKeysetsWithToken calls the broker server to get available keysets using a token
func (c *VoidkeyClient) GetAvailableKeysetsWithToken(token string) (map[string]map[string]string, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/keysets?token=%s", c.serverURL, token)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keysets map[string]map[string]string
	if err := json.Unmarshal(body, &keysets); err != nil {
		return nil, fmt.Errorf("failed to parse keysets response: %w", err)
	}

	return keysets, nil
}

// GetKeysetKeysWithToken calls the broker server to get keys for a specific keyset using a token
func (c *VoidkeyClient) GetKeysetKeysWithToken(token string, keysetName string) (map[string]string, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/keysets/keys?token=%s&keyset=%s", c.serverURL, token, keysetName)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keys map[string]string
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, fmt.Errorf("failed to parse keys response: %w", err)
	}

	return keys, nil
}

// New key-based methods

// MintKeys calls the broker server to mint specific keys using the new API
func (c *VoidkeyClient) MintKeys(oidcToken string, idpName string, keys []string, duration int, all bool) (map[string]KeyCredentialResponse, error) {
	// Prepare request
	reqBody := struct {
		OidcToken string   `json:"oidcToken"`
		IdpName   string   `json:"idpName,omitempty"`
		Keys      []string `json:"keys,omitempty"`
		Duration  int      `json:"duration,omitempty"`
		All       bool     `json:"all,omitempty"`
	}{
		OidcToken: oidcToken,
		IdpName:   idpName,
		Keys:      keys,
		Duration:  duration,
		All:       all,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to new keys endpoint
	url := fmt.Sprintf("%s/credentials/mint-keys", c.serverURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keyResponses map[string]KeyCredentialResponse
	if err := json.Unmarshal(body, &keyResponses); err != nil {
		return nil, fmt.Errorf("failed to parse key responses: %w", err)
	}

	return keyResponses, nil
}

// GetAvailableKeys calls the broker server to get available keys for an identity
func (c *VoidkeyClient) GetAvailableKeys(token string) ([]string, error) {
	// Make HTTP request
	url := fmt.Sprintf("%s/credentials/keys?token=%s", c.serverURL, token)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to broker server at %s: %w", c.serverURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var keys []string
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, fmt.Errorf("failed to parse keys response: %w", err)
	}

	return keys, nil
}

// KeyCredentialResponse represents a single key's credential response
type KeyCredentialResponse struct {
	Credentials map[string]string `json:"credentials"`
	ExpiresAt   string            `json:"expiresAt"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}
