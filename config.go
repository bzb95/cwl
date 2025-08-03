package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDir  = ".config/cwlogs"
	configFile = "config.json"
)

type Config struct {
	LogGroup string `json:"log_group"`
	Profile  string `json:"profile"`
	Region   string `json:"region"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configDir, configFile), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // Config doesn't exist yet
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func runSetup() error {
	var config Config

	fmt.Print("Enter CloudWatch log group name: ")
	fmt.Scanln(&config.LogGroup)

	fmt.Print("Enter AWS profile name (default: default): ")
	fmt.Scanln(&config.Profile)

	fmt.Print("Enter AWS region (default: us-west-2): ")
	fmt.Scanln(&config.Region)

	if config.Profile == "" {
		config.Profile = "default"
	}

	if config.Region == "" {
		config.Region = "us-west-2"
	}

	if config.LogGroup == "" {
		return fmt.Errorf("log group name cannot be empty")
	}

	return saveConfig(&config)
}
