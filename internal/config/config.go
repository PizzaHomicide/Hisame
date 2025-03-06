package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Auth   AuthConfig   `yaml:"auth"`
	Player PlayerConfig `yaml:"player"`
	UI     UIConfig     `yaml:"ui"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Token string `yaml:"token,omitempty"`
}

// PlayerConfig contains media player settings
type PlayerConfig struct {
	Type string `yaml:"type"` // "mpv", "custom"
	Path string `yaml:"path"`
	Args string `yaml:"args"`
}

// UIConfig contains UI display preferences
type UIConfig struct {
	TitleLanguage string `yaml:"title_language"`
}

func Load() (*Config, error) {
	config, err := loadFromDisk()
	if err != nil {
		return nil, err
	}
	applyEnvVarOverrides(config)
	return config, nil
}

// Load loads the configuration from disk, creating a default configuration file if none exists
func loadFromDisk() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create a default config
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		defaultConfig := createDefaultConfig()
		// If there is an error saving the default config, then still let the application startup using the defaults.
		_ = defaultConfig.Save()
		return defaultConfig, nil
	}

	// Read config from disk
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse the config yaml
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config dir if not exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// getConfigPath returns the path to the config file.  Uses the environment variable override if present, else tries
// to use OS config location defaults.
func getConfigPath() (string, error) {
	configPath := os.Getenv("HISAME_CONFIG_PATH")
	if configPath != "" {
		return configPath, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	hisameConfigDir := filepath.Join(configDir, "hisame")
	return filepath.Join(hisameConfigDir, "config.yaml"), nil
}

// createDefaultConfig creates a config with all default values
func createDefaultConfig() *Config {
	return &Config{
		Auth: AuthConfig{},
		Player: PlayerConfig{
			Type: "mpv",
		},
		UI: UIConfig{
			TitleLanguage: "english",
		},
	}
}
