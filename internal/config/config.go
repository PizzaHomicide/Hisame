package config

import (
	"dario.cat/mergo"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the application configuration
type Config struct {
	Auth    AuthConfig    `yaml:"auth,omitempty"`
	Player  PlayerConfig  `yaml:"player,omitempty"`
	UI      UIConfig      `yaml:"ui,omitempty"`
	Logging LoggingConfig `yaml:"logging,omitempty"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Token string `yaml:"token,omitempty,omitempty"`
}

// PlayerConfig contains media player settings
type PlayerConfig struct {
	Type            string `yaml:"type,omitempty"` // "mpv", "custom"
	Path            string `yaml:"path,omitempty"`
	Args            string `yaml:"args,omitempty"`
	TranslationType string `yaml:"translation_type,omitempty"` // "sub", "dub"
}

// UIConfig contains UI display preferences
type UIConfig struct {
}

// LoggingConfig contains log related settings
type LoggingConfig struct {
	Level    string `yaml:"level,omitempty"`
	FilePath string `yaml:"file_path,omitempty"`
}

// Load builds a configuration struct from multiple sources using these steps:
// 1. Create a base config with default values
// 2. If no config file exists on disk, save the default config to that location
// 3. Apply 'dynamic' properties.  Dynamic properties are those that are determined at runtime, for example log file location which is different per OS.
// 4. Load & merge the config file, overwriting any defaults with user-specified values
// 5. Apply environment variable overrides
func Load() (*Config, error) {
	// 1. Start with base defaults
	cfg := createBaseDefaultConfig()

	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("unable to determine config file path: %w", err)
	}

	// 2. If no config file exists on disk, then write a default one
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		// If there is an error saving the default config, then still let the application startup using the defaults.
		_ = save(cfg, configPath)
	}

	// 3. Apply dynamic defaults if necessary
	applyDynamicDefaults(cfg)

	// 4. Load the config from disk and merge it into the base defaults
	fileConfig, err := loadFromDisk(configPath)
	if err != nil {
		return nil, err
	}
	// Overrides the config with any values coming from the loaded file
	if err = mergo.Merge(cfg, fileConfig, mergo.WithOverride); err != nil {
		return nil, fmt.Errorf("error merging config loaded from disk: %w", err)
	}

	// 5. Apply the environment variable overrides which take precedence
	applyEnvVarOverrides(cfg)

	return cfg, nil
}

// applyDynamicDefaults sets runtime-determined default values for any properties that haven't been explicitly configured.
// Unlike static defaults, these values might change between runs based on the environment or system configuration.
func applyDynamicDefaults(cfg *Config) {
	cfg.Logging.FilePath = defaultLogFilePath()
}

// loadFromDisk loads the YAML config from disk and returns the unmarshalled Config
func loadFromDisk(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %w", err)
	}

	return cfg, nil
}

func save(cfg *Config, configPath string) error {
	// Create config dir if not exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// UpdateConfig reads the existing config, applies the update function, and saves it back to disk
func UpdateConfig(updateFn func(*Config)) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("unable to determine config file path: %w", err)
	}

	cfg, err := loadFromDisk(configPath)
	if err != nil {
		return fmt.Errorf("error loading config file from disk: %w", err)
	}

	// Apply the updates
	updateFn(cfg)

	return save(cfg, configPath)
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
func createBaseDefaultConfig() *Config {
	return &Config{
		Auth: AuthConfig{},
		Player: PlayerConfig{
			Type:            "mpv",
			Path:            "mpv",
			TranslationType: "sub",
		},
		UI: UIConfig{},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

// defaultLogFilePath returns the path to the log file.  Tries to use expected OS location defaults.
func defaultLogFilePath() string {
	var basePath string
	homedir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to logging in the current directory if home directory cannot be determined
		return filepath.Join(".", "hisame.log")
	}

	switch runtime.GOOS {
	case "windows":
		// Windows:  %LOCALAPPDATA%\hisame\logs
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			basePath = filepath.Join(appData, "hisame", "logs")
		} else {
			basePath = filepath.Join(homedir, "AppData", "local", "hisame", "logs")
		}
	case "darwin":
		// macOS:  ~/Library/Logs/hisame
		basePath = filepath.Join(homedir, "Library", "Logs", "hisame")
	default:
		// Linux/BSD:  XDG_STATE_HOME
		if xdgState := os.Getenv("XDG_STATE_HOME"); xdgState != "" {
			basePath = filepath.Join(xdgState, "hisame", "logs")
		} else {
			basePath = filepath.Join(homedir, ".local", "state", "hisame", "logs")
		}
	}

	err = os.MkdirAll(basePath, 0700)
	if err != nil {
		// If we failed to create the directory, fallback to logging in the current directory
		return filepath.Join(".", "hisame.log")
	}
	return filepath.Join(basePath, "hisame.log")
}
