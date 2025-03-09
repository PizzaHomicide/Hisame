package config

import (
	"os"
)

type envVar struct {
	name  string
	desc  string
	apply func(*Config, string)
}

var supportedEnvVars = []envVar{
	{
		// Only here for documentation purposes.  Does not override any values in the config as this environment variable
		// points to where the config should be loaded.  It is handled prior to loading the config.
		name:  "HISAME_CONFIG_PATH",
		desc:  "Sets the path to the config file.  Default: OS-specific config directory",
		apply: func(c *Config, s string) {}, // Special case, no-op
	},
	{
		name:  "HISAME_CONFIG_AUTH_TOKEN",
		desc:  "Set the AniList authentication token.  Default: None",
		apply: func(c *Config, s string) { c.Auth.Token = s },
	},
	{
		name:  "HISAME_CONFIG_UI_TITLE_LANGUAGE",
		desc:  "Sets the preferred title language for displaying anime titles.  Default: english",
		apply: func(c *Config, s string) { c.UI.TitleLanguage = s },
	},
	{
		name:  "HISAME_CONFIG_PLAYER_TYPE",
		desc:  "Sets the video player type.  Should be one of `mpv` or `custom`.  Default: mpv",
		apply: func(c *Config, s string) { c.Player.Type = s },
	},
	{
		name:  "HISAME_CONFIG_PLAYER_PATH",
		desc:  "Sets the path to a video player binary.  Default: mpv",
		apply: func(c *Config, s string) { c.Player.Path = s },
	},
	{
		name:  "HISAME_CONFIG_PLAYER_ARGS",
		desc:  "Sets the path to a video player argument.  Default: None",
		apply: func(c *Config, s string) { c.Player.Args = s },
	},
	{
		name:  "HISAME_CONFIG_LOGGING_LEVEL",
		desc:  "Sets the logging level.  One of: debug, info, warn, error.  Default: info",
		apply: func(c *Config, s string) { c.Logging.Level = s },
	},
	{
		name:  "HISAME_CONFIG_LOGGING_FILE_PATH",
		desc:  "Sets the logging file path.  Default: OS-specific",
		apply: func(c *Config, s string) { c.Logging.FilePath = s },
	},
}

func applyEnvVarOverrides(c *Config) {
	for _, envVar := range supportedEnvVars {
		if value := os.Getenv(envVar.name); value != "" {
			envVar.apply(c, value)
		}
	}
}
