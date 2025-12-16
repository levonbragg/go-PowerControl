package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Username        string `json:"username"`
	PasswordHash    string `json:"passwordHash"`
	MQTTServer      string `json:"mqttServer"`
	ServerPort      int    `json:"serverPort"`
	SubscribeString string `json:"subscribeString"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		ServerPort:      1883,
		SubscribeString: "power/#",
	}
}

// getConfigPath returns the OS-specific configuration file path
func getConfigPath() (string, error) {
	var configDir string

	// Determine config directory based on OS
	if os.Getenv("APPDATA") != "" {
		// Windows
		configDir = filepath.Join(os.Getenv("APPDATA"), "GoMQTTPowerControl")
	} else {
		// Linux/Unix
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config", "go-mqtt-power-control")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the configuration from disk
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Validate before saving
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with restricted permissions (user read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", c.ServerPort)
	}

	if c.SubscribeString == "" {
		c.SubscribeString = "power/#"
	}

	return nil
}

// IsEmpty checks if the config has required fields set
func (c *Config) IsEmpty() bool {
	return c.MQTTServer == "" || c.Username == ""
}

// SetPassword encrypts and stores the password
func (c *Config) SetPassword(plaintext string) error {
	encrypted, err := EncryptPassword(plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	c.PasswordHash = encrypted
	return nil
}

// GetPassword decrypts and returns the password
func (c *Config) GetPassword() (string, error) {
	if c.PasswordHash == "" {
		return "", nil
	}

	plaintext, err := DecryptPassword(c.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}
	return plaintext, nil
}
