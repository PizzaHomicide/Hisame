package player

import (
	"context"
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
						Type: PlaybackEnded,
					}
					return
				}

				// Process events - in the future, we could handle property changes to track progress
				if event.Event == "end-file" {
					log.Info("MPV playback ended")
					events <- PlaybackEvent{
						Type: PlaybackEnded,
					}
					return
				}
			}
		}
	}()

	return events, nil
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
