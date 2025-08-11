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

// New key-based client method tests

func TestVoidkeyClient_MintKeys_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID":     "AKIAMINIO123",
				"MINIO_SECRET_ACCESS_KEY": "miniosecret123",
				"MINIO_ENDPOINT":          "http://localhost:9000",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
			Metadata:  map[string]any{"provider": "minio-test", "keyName": "MINIO_CREDENTIALS"},
		},
		"AWS_CREDENTIALS": {
			Credentials: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAAWS123",
				"AWS_SECRET_ACCESS_KEY": "awssecret123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
			Metadata:  map[string]any{"provider": "aws-test", "keyName": "AWS_CREDENTIALS"},
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	keys := []string{"MINIO_CREDENTIALS", "AWS_CREDENTIALS"}
	keyResponses, err := client.MintKeys("test-token", "test-idp", keys, 1800, false)

	assert.NoError(t, err)
	assert.NotNil(t, keyResponses)
	assert.Len(t, keyResponses, 2)
	assert.Contains(t, keyResponses, "MINIO_CREDENTIALS")
	assert.Contains(t, keyResponses, "AWS_CREDENTIALS")
	assert.Equal(t, "AKIAMINIO123", keyResponses["MINIO_CREDENTIALS"].Credentials["MINIO_ACCESS_KEY_ID"])
	assert.Equal(t, "AKIAAWS123", keyResponses["AWS_CREDENTIALS"].Credentials["AWS_ACCESS_KEY_ID"])

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_MintKeys_AllFlag(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeyResponses := map[string]KeyCredentialResponse{
		"MINIO_CREDENTIALS": {
			Credentials: map[string]string{
				"MINIO_ACCESS_KEY_ID": "AKIAMINIO123",
			},
			ExpiresAt: "2025-01-01T12:00:00Z",
		},
	}

	responseBody, _ := json.Marshal(expectedKeyResponses)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	keyResponses, err := client.MintKeys("test-token", "test-idp", nil, 0, true)

	assert.NoError(t, err)
	assert.NotNil(t, keyResponses)
	assert.Len(t, keyResponses, 1)
	assert.Contains(t, keyResponses, "MINIO_CREDENTIALS")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_MintKeys_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Key not found")),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	keys := []string{"INVALID_KEY"}
	keyResponses, err := client.MintKeys("test-token", "test-idp", keys, 0, false)

	assert.Error(t, err)
	assert.Nil(t, keyResponses)
	assert.Contains(t, err.Error(), "server returned error 500")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_MintKeys_InvalidJSON(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}

	mockClient.On("Post", "http://localhost:3000/credentials/mint", "application/json", mock.Anything).Return(resp, nil)

	keys := []string{"MINIO_CREDENTIALS"}
	keyResponses, err := client.MintKeys("test-token", "test-idp", keys, 0, false)

	assert.Error(t, err)
	assert.Nil(t, keyResponses)
	assert.Contains(t, err.Error(), "failed to parse key responses")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_GetAvailableKeys_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	expectedKeys := []string{"MINIO_CREDENTIALS", "AWS_CREDENTIALS", "GCP_CREDENTIALS"}

	responseBody, _ := json.Marshal(expectedKeys)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/keys?token=test-token").Return(resp, nil)

	keys, err := client.GetAvailableKeys("test-token")

	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Len(t, keys, 3)
	assert.Equal(t, expectedKeys, keys)

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_GetAvailableKeys_EmptyResponse(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	responseBody, _ := json.Marshal([]string{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/keys?token=test-token").Return(resp, nil)

	keys, err := client.GetAvailableKeys("test-token")

	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Len(t, keys, 0)

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_GetAvailableKeys_ServerError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("Invalid token")),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/keys?token=invalid-token").Return(resp, nil)

	keys, err := client.GetAvailableKeys("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "server returned error 500")

	mockClient.AssertExpectations(t)
}

func TestVoidkeyClient_GetAvailableKeys_InvalidJSON(t *testing.T) {
	mockClient := &MockHTTPClient{}
	client := NewVoidkeyClient(mockClient, "http://localhost:3000")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}

	mockClient.On("Get", "http://localhost:3000/credentials/keys?token=test-token").Return(resp, nil)

	keys, err := client.GetAvailableKeys("test-token")

	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "failed to parse keys response")

	mockClient.AssertExpectations(t)
}
