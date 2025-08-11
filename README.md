# Voidkey CLI

A Go-based command-line client for interacting with the Voidkey zero-trust credential broker system.

## Overview

The Voidkey CLI provides a simple interface for requesting temporary credentials from cloud providers through the Voidkey broker. It handles OIDC token authentication and credential requests, making it easy to integrate secure credential management into scripts and automation workflows.

## Architecture

The CLI participates in the following zero-trust credential broker workflow:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │──▶│  Client IdP │    │   Voidkey   │──▶│  Broker IdP │──▶│   Access    │
│     CLI     │    │  (Auth0,    │    │   Broker    │    │ (Keycloak,  │    │  Provider   │
│             │    │ GitHub, etc)│    │             │    │  Okta, etc) │    │    (STS)    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │                   │
       │ 1. Get client     │                   │                   │                   │
       │    OIDC token     │                   │                   │                   │
       │◀─────────────────│                   │                   │                   │
       │                                       │                   │                   │
       │ 2. Request credentials with token     │                   │                   │
       │─────────────────────────────────────▶│                   │                   │
       │                                       │                   │                   │
       │                             3. Validate client token      │                   │
       │                                       │                   │                   │
       │                                       │ 4. Get broker     │                   │
       │                                       │    OIDC token     │                   │
       │                                       │◀─────────────────│                   │
       │                                       │                                       │
       │                                       │ 5. Mint credentials with broker token │
       │                                       │─────────────────────────────────────▶│
       │                                       │                                       │
       │                                       │ 6. Return temp credentials            │
       │                                       │◀─────────────────────────────────────│
       │                                       │                                       │
       │ 7. Return temp credentials to client  │                                       │
       │◀─────────────────────────────────────│                                       │
       │                                                                               │
       │ 8. Use credentials for operations                                             │
       │─────────────────────────────────────────────────────────────────────────────▶│
```

This true zero-trust architecture ensures:
- **Client Authentication**: Client authenticates with its own IdP (Auth0, GitHub Actions, etc.)
- **Broker Authentication**: Broker independently authenticates with its own IdP to access providers
- **Token Validation**: Broker validates client tokens without sharing credentials
- **Credential Isolation**: No shared secrets between client and access provider

## Installation

### Build from Source

```bash
go build -o voidkey main.go
```

### Install

```bash
# Make executable and add to PATH
chmod +x voidkey
sudo mv voidkey /usr/local/bin/
```

## Usage

### Basic Command Structure

```bash
voidkey --broker-url <broker-url> --token <oidc-token> --keys <key-names>
```

### Parameters

- `--broker-url`: URL of the Voidkey broker server
- `--token`: OIDC token for authentication (from your IdP)
- `--keys`: Comma-separated list of key names to request credentials for
- `--keyset`: (Legacy) Keyset name for batch credential requests

### Examples

#### Request credentials for specific keys

```bash
voidkey --broker-url https://broker.example.com \
        --token eyJhbGciOiJSUzI1NiIs... \
        --keys s3-readonly,s3-readwrite
```

#### List available keys for your identity

```bash
voidkey --broker-url https://broker.example.com \
        --token eyJhbGciOiJSUzI1NiIs... \
        --list-keys
```

#### Use with environment variables

```bash
export VOIDKEY_BROKER_URL=https://broker.example.com
export VOIDKEY_TOKEN=eyJhbGciOiJSUzI1NiIs...

voidkey --keys s3-readonly
```

### Output Formats

The CLI supports multiple output formats for credentials:

- **JSON**: Complete credential structure
- **Environment Variables**: Export statements for shell integration
- **AWS Credentials File**: AWS CLI compatible format

## Integration Examples

### Shell Script Integration

```bash
#!/bin/bash

# Get credentials and set environment variables
eval $(voidkey --keys s3-readonly --format env)

# Use AWS CLI with temporary credentials
aws s3 ls s3://my-bucket/
```

### CI/CD Pipeline Integration

```yaml
steps:
  - name: Get AWS Credentials
    run: |
      voidkey --keys ci-deployment --format json > /tmp/creds.json
      echo "AWS_ACCESS_KEY_ID=$(jq -r '.AccessKeyId' /tmp/creds.json)" >> $GITHUB_ENV
      echo "AWS_SECRET_ACCESS_KEY=$(jq -r '.SecretAccessKey' /tmp/creds.json)" >> $GITHUB_ENV
      echo "AWS_SESSION_TOKEN=$(jq -r '.SessionToken' /tmp/creds.json)" >> $GITHUB_ENV
```

## Development

### Running Tests

```bash
go test ./...
```

### Testing Specific Packages

```bash
go test ./cmd
```

### Building for Multiple Platforms

```bash
GOOS=linux GOARCH=amd64 go build -o voidkey-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o voidkey-darwin-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o voidkey-windows-amd64.exe main.go
```

## Configuration

### Environment Variables

- `VOIDKEY_BROKER_URL`: Default broker URL
- `VOIDKEY_TOKEN`: Default OIDC token
- `VOIDKEY_DEBUG`: Enable debug logging

### Configuration File

Create `~/.voidkey/config.yaml`:

```yaml
broker_url: https://broker.example.com
default_keys:
  - s3-readonly
timeout: 30s
```

## Troubleshooting

### Common Issues

1. **Invalid token**: Ensure your OIDC token is valid and not expired
2. **Network connectivity**: Verify the broker URL is accessible
3. **Permission denied**: Check that your identity has access to the requested keys
4. **Token format**: Ensure the token includes required claims (subject, audience, etc.)

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
VOIDKEY_DEBUG=true voidkey --keys s3-readonly
```

## Security Considerations

- OIDC tokens should be handled securely and not logged
- Temporary credentials automatically expire for security
- Always use HTTPS for broker communication
- Store tokens in secure locations (environment variables, secret managers)
- Rotate OIDC tokens regularly according to your identity provider's recommendations
- Client and broker use separate IdPs for true zero-trust architecture
