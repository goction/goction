package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Config holds the application configuration
type Config struct {
	GoctionsDir string `json:"goctions_dir"`
	Port        int    `json:"port"`
	LogFile     string `json:"log_file"`
	APIToken    string `json:"api_token"`
	StatsFile   string `json:"stats_file"`
}

// Load reads the configuration file and returns a Config struct
func Load() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig(configPath)
	}

	return loadExistingConfig(configPath)
}

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

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
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	// Remove the existing config file
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing config file: %w", err)
	}

	// Create a new default config
	_, err = createDefaultConfig(configPath)
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

// Helper functions

func getConfigDir() (string, error) {
	if os.Geteuid() == 0 {
		return "/etc/goction", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "goction"), nil
}

func createDefaultConfig(configPath string) (*Config, error) {
	cfg := &Config{
		GoctionsDir: filepath.Join(filepath.Dir(configPath), "goctions"),
		Port:        8080,
		LogFile:     filepath.Join(filepath.Dir(configPath), "goction.log"),
		APIToken:    uuid.New().String(),
		StatsFile:   filepath.Join(filepath.Dir(configPath), "goction_stats.json"),
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