package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestConfig(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "hisame-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	})

	tmpConfigPath := filepath.Join(tmpDir, "config.yaml")
	setEnv(t, "HISAME_CONFIG_PATH", tmpConfigPath)

	t.Cleanup(func() {
		cleanupEnvVars(t)
	})

	return tmpConfigPath
}

// TestConfigIntegration tests the config package with actual file operations
// This test uses a temporary directory to avoid interfering with real user configs
func TestConfigIntegration(t *testing.T) {
	// Test loading when no config exists (should create default)
	t.Run("LoadDefaultConfig", func(t *testing.T) {
		tmpConfigPath := setupTestConfig(t)
		config := loadConfig(t)

		// Verify default values
		assert.Equal(t, "mpv", config.Player.Type)
		assert.Equal(t, "sub", config.Player.TranslationType)
		assert.Equal(t, "info", config.Logging.Level)
		assert.NotEmpty(t, config.Logging.FilePath)

		// Verify file was created
		if _, err := os.Stat(tmpConfigPath); os.IsNotExist(err) {
			t.Errorf("Config file was not created at %s", tmpConfigPath)
		}

		// Load the file from disk to assert that the 'dynamic' configurations were not saved when the default config was written
		savedConfig, _ := loadFromDisk(tmpConfigPath)
		assert.Empty(t, savedConfig.Logging.FilePath)
	})

	// Test saving and loading custom values
	t.Run("SaveAndLoadConfig", func(t *testing.T) {
		tmpConfigPath := setupTestConfig(t)
		// Create a config with custom values
		customConfig := &Config{
			Auth: AuthConfig{
				Token: "test-token",
			},
			Player: PlayerConfig{
				Type:            "custom",
				Path:            "/usr/bin/vlc",
				Args:            "--fullscreen",
				TranslationType: "dub",
			},
			UI: UIConfig{},
			Logging: LoggingConfig{
				Level:    "error",
				FilePath: "/var/log/hisame.log",
			},
		}

		saveConfig(t, customConfig, tmpConfigPath)
		loadedConfig := loadConfig(t)

		// Verify loaded values match what we saved
		assert.Equal(t, "test-token", loadedConfig.Auth.Token)
		assert.Equal(t, "custom", loadedConfig.Player.Type)
		assert.Equal(t, "/usr/bin/vlc", loadedConfig.Player.Path)
		assert.Equal(t, "--fullscreen", loadedConfig.Player.Args)
		assert.Equal(t, "dub", loadedConfig.Player.TranslationType)
		assert.Equal(t, "error", loadedConfig.Logging.Level)
		assert.Equal(t, "/var/log/hisame.log", loadedConfig.Logging.FilePath)
	})

	// Test invalid YAML handling
	t.Run("InvalidConfig", func(t *testing.T) {
		tmpConfigPath := setupTestConfig(t)
		// Write invalid YAML to the config file
		if err := os.WriteFile(tmpConfigPath, []byte("invalid: yaml: ["), 0600); err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		// Attempt to load the invalid config
		_, err := Load()
		if err == nil {
			t.Error("Expected error when loading invalid YAML, got nil")
		}
	})

	t.Run("EnvironmentVariableOverrides", func(t *testing.T) {
		setupTestConfig(t)

		setEnv(t, "HISAME_CONFIG_AUTH_TOKEN", "test-token")
		setEnv(t, "HISAME_CONFIG_PLAYER_TYPE", "custom")
		setEnv(t, "HISAME_CONFIG_PLAYER_PATH", "/vlc")
		setEnv(t, "HISAME_CONFIG_PLAYER_ARGS", "--fullscreen")
		setEnv(t, "HISAME_CONFIG_PLAYER_TRANSLATION_TYPE", "dub")
		setEnv(t, "HISAME_CONFIG_LOGGING_LEVEL", "warn")
		setEnv(t, "HISAME_CONFIG_LOGGING_FILE_PATH", "/hisame.log")

		config := loadConfig(t)

		assert.Equal(t, "test-token", config.Auth.Token)
		assert.Equal(t, "custom", config.Player.Type)
		assert.Equal(t, "/vlc", config.Player.Path)
		assert.Equal(t, "--fullscreen", config.Player.Args)
		assert.Equal(t, "dub", config.Player.TranslationType)
		assert.Equal(t, "warn", config.Logging.Level)
		assert.Equal(t, "/hisame.log", config.Logging.FilePath)

		// Remove the HISAME_CONFIG_UI_TITLE_LANGUAGE env var, then reload the config.
		// This ensures that the env var overrides were not persisted to disk.
		unsetEnv(t, "HISAME_CONFIG_LOGGING_LEVEL")

		config = loadConfig(t)

		assert.Equal(t, "info", config.Logging.Level)
	})

	t.Run("ModifyConfig", func(t *testing.T) {
		setupTestConfig(t)
		config := loadConfig(t)

		assert.Equal(t, "mpv", config.Player.Type)

		err := UpdateConfig(func(config *Config) {
			config.Player.Type = "custom"
		})
		if err != nil {
			t.Fatalf("Failed to update config: %v", err)
		}

		// Reload the config and ensure it has the new value
		config = loadConfig(t)
		assert.Equal(t, "custom", config.Player.Type)
	})
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	err := os.Unsetenv(key)
	if err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}
}

func saveConfig(t *testing.T, config *Config, configPath string) {
	t.Helper()
	if err := save(config, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
}

func loadConfig(t *testing.T) *Config {
	t.Helper()
	config, err := Load()
	if err != nil {
		t.Fatalf("Loading of config failed: %v", err)
	}
	return config
}

// Removes any env vars with the HISAME_CONFIG prefix to ensure test isolation
func cleanupEnvVars(t *testing.T) {
	t.Helper()

	for _, envVar := range os.Environ() {
		if key := strings.Split(envVar, "=")[0]; strings.HasPrefix(key, "HISAME_CONFIG") {
			unsetEnv(t, key)
		}
	}
}
