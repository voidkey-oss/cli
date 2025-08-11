package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the persistent CLI configuration
type Config struct {
	Server   string `yaml:"server"`
	IdpName  string `yaml:"idp_name,omitempty"`
	TokenEnv string `yaml:"token_env,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server:   "http://localhost:3000",
		TokenEnv: "OIDC_TOKEN",
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".voidkey")
	return filepath.Join(configDir, "config.yaml"), nil
}

// LoadConfig loads configuration from file, or returns default if file doesn't exist
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing fields
	if config.Server == "" {
		config.Server = "http://localhost:3000"
	}
	if config.TokenEnv == "" {
		config.TokenEnv = "OIDC_TOKEN"
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetToken retrieves the OIDC token from the configured environment variable or fallbacks
func (c *Config) GetToken() string {
	token, _ := c.GetTokenWithSource()
	return token
}

// GetTokenWithSource retrieves the OIDC token and returns both the token and the env var name that provided it
func (c *Config) GetTokenWithSource() (string, string) {
	// Try the configured token environment variable
	if c.TokenEnv != "" {
		if token := os.Getenv(c.TokenEnv); token != "" {
			return token, c.TokenEnv
		}
	}

	// Fallback to common environment variables
	if token := os.Getenv("OIDC_TOKEN"); token != "" {
		return token, "OIDC_TOKEN"
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, "GITHUB_TOKEN"
	}

	return "", ""
}
