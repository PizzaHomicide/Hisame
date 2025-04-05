package player

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
)

// MPVPlayer implements the VideoPlayer interface for MPV
type MPVPlayer struct {
	config     *config.Config
	ipcClient  *MPVIPCClient
	cmd        *exec.Cmd
	socketPath string
}

// NewMPVPlayer creates a new MPV player instance
func NewMPVPlayer(cfg *config.Config) *MPVPlayer {
	socketPath := GetMPVSocketPath()
	return &MPVPlayer{
		config:     cfg,
		socketPath: socketPath,
		ipcClient:  NewMPVIPCClient(socketPath),
	}
}

// Play starts playback of the given URL, monitors for playback start, and returns a notification channel
func (p *MPVPlayer) Play(ctx context.Context, url string) (<-chan PlaybackEvent, error) {
	log.Info("Starting MPV playback", "url", url)

	// Create notification channel for playback events
	events := make(chan PlaybackEvent, 10)

	// Get MPV binary path from config
	mpvPath := p.config.Player.Path
	if mpvPath == "" {
		mpvPath = "mpv"
	}

	// Build the arguments
	args := []string{
		"--no-terminal",                      // Disable terminal control
		"--keep-open=no",                     // Exit when playback is complete
		"--input-ipc-server=" + p.socketPath, // Set IPC socket path
	}

	// Add any additional configured arguments
	if p.config.Player.Args != "" {
		customArgs := ParseArgs(p.config.Player.Args)
		args = append(args, customArgs...)
	}

	// Add the stream URL as the final argument
	args = append(args, url)

	// Create command
	cmd := exec.Command(mpvPath, args...)

	// Platform-specific process setup
	setupPlayerProcess(cmd)

	// Start MPV
	if err := cmd.Start(); err != nil {
		close(events)
		return events, fmt.Errorf("failed to start MPV: %w", err)
	}
	p.cmd = cmd

	// Release the process (platform-specific)
	if err := releasePlayerProcess(cmd); err != nil {
		log.Warn("Failed to release MPV process", "error", err)
	}

	// Start a goroutine to monitor playback
	go func() {
		defer close(events)

		// Allow time for MPV to create the socket
		time.Sleep(300 * time.Millisecond)

		// Wait for MPV to create its socket and establish connection
		connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Try to connect to MPV with retries
		err := p.ipcClient.WaitForConnection(connCtx, 20, 500*time.Millisecond)
		if err != nil {
			log.Error("Failed to connect to MPV", "error", err)
			events <- PlaybackEvent{
				Type:  PlaybackError,
				Error: err,
			}
			return
		}

		// Wait for playback to actually start
		playbackCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err = p.ipcClient.WaitForPlaybackStart(playbackCtx, 30*time.Second)
		if err != nil {
			log.Error("Failed to detect MPV playback start", "error", err)
			events <- PlaybackEvent{
				Type:  PlaybackError,
				Error: err,
			}
			return
		}

		// Playback has started
		events <- PlaybackEvent{
			Type: PlaybackStarted,
		}

		var playbackTime, duration float64
		// Used for logging.  We want to log out progress updates infrequently and will be casting a float to an int,
		// so will get many events for the same percentage number - therefore we need to track the last logged number
		// so we don't spam logs of that one number
		var lastLoggedProgress int = -1

		// Keep processing events until MPV exits or context is cancelled
		mpvEventCh := p.ipcClient.Events()
		for {
			select {
			case <-ctx.Done():
				log.Debug("Context cancelled, stopping MPV monitoring")
				return
			case event, ok := <-mpvEventCh:
				if !ok {
					log.Debug("MPV event channel closed")
					events <- PlaybackEvent{
						Type:     PlaybackEnded,
						Progress: p.calculateProgressPercentage(playbackTime, duration),
					}
					return
				}

				// Process events - in the future, we could handle property changes to track progress
				if event.Event == "end-file" {
					log.Info("MPV playback ended")
					events <- PlaybackEvent{
						Type:     PlaybackEnded,
						Progress: p.calculateProgressPercentage(playbackTime, duration),
					}
					return
				}
				if event.Event == "property-change" {
					if durationValue, err := p.extractEventDataFloat(event, "duration"); err == nil {
						log.Trace("Setting video duration", "duration", durationValue)
						duration = durationValue
					}
					if playbackValue, err := p.extractEventDataFloat(event, "playback-time"); err == nil {
						log.Trace("Setting playback time", "playback-time", playbackValue)
						playbackTime = playbackValue

						progress := int(p.calculateProgressPercentage(playbackTime, duration))
						if progress != lastLoggedProgress && (progress%5 == 0 || absInt(lastLoggedProgress-progress) >= 5) {
							log.Info("Playback progress", "percent", progress)
							lastLoggedProgress = progress
						}
					}
				}
			}
		}
	}()

	return events, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (p *MPVPlayer) extractEventDataFloat(event MPVEvent, targetName string) (float64, error) {
	if event.Name != targetName {
		return 0.0, fmt.Errorf("event name %s does not match target name %s", event.Name, targetName)
	}

	var value float64
	if err := json.Unmarshal(event.Data, &value); err != nil {
		log.Warn("Failed to unmarshal event data", "data", string(event.Data))
		return 0.0, fmt.Errorf("failed to unmarshal event data: %w", err)
	} else {
		log.Trace("Parsed value", "value", value, "name", targetName)
		return value, nil
	}
}

func (p *MPVPlayer) calculateProgressPercentage(playbackTime, duration float64) float64 {
	log.Trace("Calculating progress percentage..", "playbackTime", playbackTime, "duration", duration)
	if playbackTime == 0.0 || duration == 0.0 {
		return 0.0
	}
	return (playbackTime / duration) * 100
}

// Stop stops playback if it's active
func (p *MPVPlayer) Stop() error {
	// Close IPC connection if it exists
	if p.ipcClient != nil {
		p.ipcClient.Close()
	}

	// Kill MPV process if it exists
	if p.cmd != nil && p.cmd.Process != nil {
		log.Info("Stopping MPV playback")
		return p.cmd.Process.Kill()
	}

	return nil
}

// Cleanup performs any necessary cleanup
func (p *MPVPlayer) Cleanup() {
	p.Stop()

	// Remove socket file if it exists (Unix only)
	if _, err := os.Stat(p.socketPath); err == nil {
		if err := os.Remove(p.socketPath); err != nil {
			log.Warn("Failed to remove MPV socket file", "path", p.socketPath, "error", err)
		}
	}
}
