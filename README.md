# Hisame (氷雨)

Hisame is a terminal-based AniList client that helps you manage your anime list and watch episodes directly from the terminal.

> **Note:** Hisame is currently in alpha. Expect bugs and changes as development continues.

## Features

- Authentication with AniList
- Browse and filter your anime list
- Watch episodes directly from the application
- Automatic progress tracking with MPV integration

## Installation

1. Download the latest release from [GitHub Releases](https://github.com/PizzaHomicide/hisame/releases)
2. Extract the binary for your platform
3. Run the extracted `hisame` executable from a terminal

### MPV Requirement

Hisame requires [MPV](https://mpv.io/) for media playback and automatic progress tracking.

#### Installing MPV:

- **Windows**: Install using [winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/):
  ```
  winget install mpv
  ```
  Alternatively, download an MPV binary from the [MPV website](https://mpv.io), and then read the configuration information below to understand how to set the player path configuration, and put the full path to the binary in here.


- **macOS**: Install using [Homebrew](https://brew.sh/):
  ```
  brew install mpv
  ```

- **Linux**: Install using your distribution's package manager:
  ```
  # Debian/Ubuntu
  sudo apt install mpv
  
  # Fedora
  sudo dnf install mpv
  
  # Arch Linux
  sudo pacman -S mpv
  ```

## Configuration

Hisame uses a YAML configuration file located at:

- **Windows**: `%APPDATA%%\hisame\config.yaml`
- **macOS**: `~/Library/Application Support/hisame/config.yaml`
- **Linux**: `$XDG_CONFIG_HOME/hisame/config.yaml` (or `~/.config/hisame/config.yaml` if XDG_CONFIG_HOME is not set)

The config file will be created automatically on first run with default values.

### Configuration Options

```yaml
auth:
  token: ""        # AniList authentication token (managed by Hisame)
player:
  type: "mpv"      # Player type (mpv or custom)
  path: "mpv"      # Path to media player executable
  args: ""         # Additional arguments to pass to the player
  translation_type: "sub"  # Preferred translation type (sub or dub)
logging:
  level: "info"    # Logging level (debug, info, warn, error)
  file_path: ""    # Path to log file (auto-generated if not specified)
```

### Log File Locations

Hisame creates log files at these default locations:

- **Windows**: `%LOCALAPPDATA%\hisame\logs\hisame.log`
- **macOS**: `~/Library/Logs/hisame/hisame.log`
- **Linux**: `$XDG_STATE_HOME/hisame/logs/hisame.log` (or `~/.local/state/hisame/logs/hisame.log` if XDG_STATE_HOME is not set)

### MPV Configuration

If MPV is not in your system PATH, you need to specify the full path to the MPV executable in the config file:

```yaml
player:
  path: "/full/path/to/mpv"  # Replace with the actual path to MPV
```

### Environment Variable Overrides

All configuration options can be overridden with environment variables:

| Environment Variable | Description |
|----------------------|-------------|
| `HISAME_CONFIG_PATH` | Path to config file |
| `HISAME_CONFIG_AUTH_TOKEN` | AniList authentication token |
| `HISAME_CONFIG_PLAYER_TYPE` | Player type (mpv or custom) |
| `HISAME_CONFIG_PLAYER_PATH` | Path to player executable |
| `HISAME_CONFIG_PLAYER_ARGS` | Additional arguments for player |
| `HISAME_CONFIG_PLAYER_TRANSLATION_TYPE` | Preferred translation type (sub or dub) |
| `HISAME_CONFIG_LOGGING_LEVEL` | Logging level |
| `HISAME_CONFIG_LOGGING_FILE_PATH` | Path to log file |

Example:
```bash
HISAME_CONFIG_PLAYER_PATH="/path/to/mpv" HISAME_CONFIG_PLAYER_TRANSLATION_TYPE="dub" ./hisame
```

## Basic Usage

When you first launch Hisame, you'll need to authenticate with AniList:

1. Press `l` to start the login process
2. A browser window will open to authenticate with AniList
3. After authentication, you'll be redirected back to Hisame

Once authenticated, you can:

- Use arrow keys to navigate the anime list
- Press `Enter` to play the next episode of selected anime
- Press `Ctrl+p` to select a specific episode to play
- Use number keys (`1-6`) to toggle status filters
- Press `/` to search your anime list
- Press `d` to view detailed information about the selected anime
- Press `+` and `-` to adjust episode progress
- Press `Ctrl+h` to access the help screen with all commands

## Limitations

- MPV is required for automatic progress tracking
- Other media players can be configured to launch videos, but progress updates will need to be done manually
- Pre-release software: expect bugs and changes

## Troubleshooting

If you encounter issues:

- Check the log file for detailed error information
- Ensure MPV is properly installed and accessible
- Verify your AniList authentication is valid
- If necessary, logout with `Ctrl+l` and re-authenticate
