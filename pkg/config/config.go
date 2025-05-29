package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDir  = ".lolarchiver"
	configFile = "config.json"
)

// Config represents the application configuration
type Config struct {
	APIKey string `json:"api_key"`
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to the config file
func Save(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, configDir)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDirPath, configFile)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetAPIKey sets the API key in the configuration
func SetAPIKey(apiKey string) error {
	config, err := Load()
	if err != nil {
		return err
	}

	config.APIKey = apiKey
	return Save(config)
}

// GetAPIKey gets the API key from the configuration
func GetAPIKey() (string, error) {
	config, err := Load()
	if err != nil {
		return "", err
	}

	if config.APIKey == "" {
		return "", fmt.Errorf("API key not set. Use 'lolarchiver-cli config set-api-key' to set it")
	}

	return config.APIKey, nil
} 