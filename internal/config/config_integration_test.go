package config

import (
	"os"
	"path/filepath"
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
		unsetEnv(t, "HISAME_CONFIG_PATH")
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
		assert(t, "mpv", config.Player.Type)
		assert(t, "english", config.UI.TitleLanguage)

		// Verify file was created
		if _, err := os.Stat(tmpConfigPath); os.IsNotExist(err) {
			t.Errorf("Config file was not created at %s", tmpConfigPath)
		}
	})

	// Test saving and loading custom values
	t.Run("SaveAndLoadConfig", func(t *testing.T) {
		setupTestConfig(t)
		// Create a config with custom values
		customConfig := &Config{
			Auth: AuthConfig{
				Token: "test-token",
			},
			Player: PlayerConfig{
				Type: "custom",
				Path: "/usr/bin/vlc",
				Args: "--fullscreen",
			},
			UI: UIConfig{
				TitleLanguage: "romaji",
			},
		}

		saveConfig(t, customConfig)
		loadedConfig := loadConfig(t)

		// Verify loaded values match what we saved
		assert(t, "test-token", loadedConfig.Auth.Token)
		assert(t, "custom", loadedConfig.Player.Type)
		assert(t, "/usr/bin/vlc", loadedConfig.Player.Path)
		assert(t, "--fullscreen", loadedConfig.Player.Args)
		assert(t, "romaji", loadedConfig.UI.TitleLanguage)
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
		setEnv(t, "HISAME_CONFIG_UI_TITLE_LANGUAGE", "romaji")
		setEnv(t, "HISAME_CONFIG_PLAYER_TYPE", "custom")
		setEnv(t, "HISAME_CONFIG_PLAYER_PATH", "/vlc")
		setEnv(t, "HISAME_CONFIG_PLAYER_ARGS", "--fullscreen")

		config := loadConfig(t)

		assert(t, "test-token", config.Auth.Token)
		assert(t, "romaji", config.UI.TitleLanguage)
		assert(t, "custom", config.Player.Type)
		assert(t, "/vlc", config.Player.Path)
		assert(t, "--fullscreen", config.Player.Args)

		// Remove the HISAME_CONFIG_UI_TITLE_LANGUAGE env var, then reload the config.
		// This ensures that the env var overrides were not persisted to disk.
		unsetEnv(t, "HISAME_CONFIG_UI_TITLE_LANGUAGE")

		config = loadConfig(t)

		assert(t, "english", config.UI.TitleLanguage)
	})
}

func assert[T comparable](t *testing.T, expected T, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
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

func saveConfig(t *testing.T, config *Config) {
	t.Helper()
	if err := config.Save(); err != nil {
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
