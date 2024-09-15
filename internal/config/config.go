package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const GoctionVersion = "1.0.0"
const ConfigDir = "/etc/goction"

// Config holds the application configuration
type Config struct {
	GoctionsDir       string `json:"goctions_dir"`
	Port              int    `json:"port"`
	LogFile           string `json:"log_file"`
	APIToken          string `json:"api_token"`
	StatsFile         string `json:"stats_file"`
	DashboardUsername string `json:"dashboard_username"`
	DashboardPassword string `json:"dashboard_password"`
}

// Load reads the configuration file and returns a Config struct
func Load() (*Config, error) {
	configPath := filepath.Join(ConfigDir, "config.json")

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configPath := filepath.Join(ConfigDir, "config.json")

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// Reset resets the configuration to default values
func Reset() error {
	configPath := filepath.Join(ConfigDir, "config.json")

	// Remove the existing config file
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing config file: %w", err)
	}

	// Create a new default config
	_, err := createDefaultConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to create new default config: %w", err)
	}

	return nil
}

// InitializeLogFile creates the log file if it doesn't exist
func (c *Config) InitializeLogFile() error {
	dir := filepath.Dir(c.LogFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the file in append mode, creating it if it doesn't exist
	file, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create or open log file: %w", err)
	}
	defer file.Close()

	return nil
}

func createDefaultConfig(configPath string) (*Config, error) {
	cfg := &Config{
		GoctionsDir:       "/etc/goction/goctions",
		Port:              8080,
		LogFile:           "/var/log/goction/goction.log",
		APIToken:          uuid.New().String(),
		StatsFile:         "/var/log/goction/goction_stats.json",
		DashboardUsername: "admin",
		DashboardPassword: uuid.New().String(),
	}

	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("failed to save default config: %w", err)
	}

	return cfg, nil
}

func loadExistingConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}
