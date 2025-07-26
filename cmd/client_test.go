package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
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

func TestNewVoidkeyClient(t *testing.T) {
	mockClient := &MockHTTPClient{}
	serverURL := "http://localhost:3000"

	client := NewVoidkeyClient(mockClient, serverURL)

	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client.client)
	assert.Equal(t, serverURL, client.serverURL)
}

func TestVoidkeyClient_MintCredentials_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedCreds := CloudCredentials{
		AccessKey:    "AKIATEST123",
		SecretKey:    "secretkey123",
		SessionToken: "sessiontoken123",
		ExpiresAt:    "2024-12-31T23:59:59Z",
	}

	responseBody, _ := json.Marshal(expectedCreds)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	credentials, err := client.MintCredentials("test-token", "test-idp")

	assert.NoError(t, err)
	assert.NotNil(t, credentials)
	assert.Equal(t, expectedCreds.AccessKey, credentials.AccessKey)
	assert.Equal(t, expectedCreds.SecretKey, credentials.SecretKey)
	assert.Equal(t, expectedCreds.SessionToken, credentials.SessionToken)
	assert.Equal(t, expectedCreds.ExpiresAt, credentials.ExpiresAt)

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_MintCredentials_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Internal server error")),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	credentials, err := client.MintCredentials("test-token", "test-idp")

	assert.Error(t, err)
	assert.Nil(t, credentials)
	assert.Contains(t, err.Error(), "server returned error 500")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_MintCredentials_InvalidJSON(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	credentials, err := client.MintCredentials("test-token", "test-idp")

	assert.Error(t, err)
	assert.Nil(t, credentials)
	assert.Contains(t, err.Error(), "failed to parse credentials response")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_ListIdpProviders_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedProviders := []IdpProvider{
		{Name: "auth0", IsDefault: true},
		{Name: "github", IsDefault: false},
		{Name: "hello-world", IsDefault: false},
	}

	responseBody, _ := json.Marshal(expectedProviders)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	providers, err := client.ListIdpProviders()

	assert.NoError(t, err)
	assert.NotNil(t, providers)
	assert.Len(t, providers, 3)
	assert.Equal(t, expectedProviders[0].Name, providers[0].Name)
	assert.Equal(t, expectedProviders[0].IsDefault, providers[0].IsDefault)

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_ListIdpProviders_EmptyResponse(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	responseBody, _ := json.Marshal([]IdpProvider{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	providers, err := client.ListIdpProviders()

	assert.NoError(t, err)
	assert.NotNil(t, providers)
	assert.Len(t, providers, 0)

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_ListIdpProviders_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Internal server error")),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	providers, err := client.ListIdpProviders()

	assert.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "server returned error 500")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_ListIdpProviders_InvalidJSON(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/idp-providers").Return(resp, nil)

	providers, err := client.ListIdpProviders()

	assert.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "failed to parse providers response")

	mockClient.AssertExpectations(t)
}
