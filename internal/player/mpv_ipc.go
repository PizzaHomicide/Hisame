package player

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/PizzaHomicide/hisame/internal/log"
)

// MPVIPCClient provides communication with a running MPV instance
type MPVIPCClient struct {
	socketPath string
	conn       net.Conn
	events     chan MPVEvent
}

// MPVEvent represents an event from MPV
type MPVEvent struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data,omitempty"`
	RequestID int             `json:"request_id,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// NewMPVIPCClient creates a new MPV IPC client
func NewMPVIPCClient(socketPath string) *MPVIPCClient {
	return &MPVIPCClient{
		socketPath: socketPath,
		events:     make(chan MPVEvent, 100),
	}
}

// GetMPVSocketPath returns the socket path for MPV IPC communication
func GetMPVSocketPath() string {
	var socketPath string

	// Use environment variable if set
	if path := os.Getenv("MPV_IPC_SOCKET"); path != "" {
		return path
	}

	// Otherwise use default location based on OS
	switch runtime.GOOS {
	case "windows":
		// Windows uses named pipes instead of unix sockets
		return `\\.\pipe\mpv-pipe`
	case "darwin":
		// macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Error("Failed to get user home directory", "error", err)
			return "/tmp/mpv-socket"
		}
		socketPath = filepath.Join(homeDir, ".config", "mpv", "socket")
	default:
		// Linux and others
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir != "" {
			socketPath = filepath.Join(runtimeDir, "mpv-socket")
		} else {
			socketPath = "/tmp/mpv-socket"
		}
	}

	return socketPath
}

// Connect establishes a connection with MPV
func (c *MPVIPCClient) Connect(ctx context.Context) error {
	// For Windows, use regular net.Dial with the named pipe
	if runtime.GOOS == "windows" {
		conn, err := net.Dial("tcp", c.socketPath)
		if err != nil {
			return fmt.Errorf("failed to connect to MPV pipe: %w", err)
		}
		c.conn = conn
		go c.readEvents()
		return nil
	}

	// For Unix systems, use Unix domain socket
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to MPV socket: %w", err)
	}

	c.conn = conn
	go c.readEvents()
	return nil
}

// WaitForConnection attempts to connect to MPV with retries
func (c *MPVIPCClient) WaitForConnection(ctx context.Context, maxAttempts int, retryDelay time.Duration) error {
	log.Debug("Waiting for MPV to create socket", "socket_path", c.socketPath, "max_attempts", maxAttempts)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if socket file exists for unix sockets
		if runtime.GOOS != "windows" {
			if _, err := os.Stat(c.socketPath); os.IsNotExist(err) {
				log.Debug("MPV socket does not exist yet", "attempt", attempt, "path", c.socketPath)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryDelay):
					continue
				}
			}
		}

		// Try to connect
		err := c.Connect(ctx)
		if err == nil {
			log.Info("Successfully connected to MPV", "attempt", attempt)
			return nil
		}

		log.Debug("Failed to connect to MPV", "attempt", attempt, "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
			// Continue and retry
		}
	}

	return fmt.Errorf("failed to connect to MPV after %d attempts", maxAttempts)
}

// Close closes the connection to MPV
func (c *MPVIPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// readEvents continuously reads events from MPV
func (c *MPVIPCClient) readEvents() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Text()

		// Log the raw data at debug level to see what MPV is sending
		log.Trace("Raw MPV event", "data", line)

		var event MPVEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			log.Error("Failed to unmarshal MPV event", "error", err)
			continue
		}

		log.Trace("Received MPV event", "event", event.Event)
		c.events <- event
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading from MPV socket", "error", err)
	}

	log.Debug("MPV event reader stopped")
	close(c.events)
}

// Events returns the channel for MPV events
func (c *MPVIPCClient) Events() <-chan MPVEvent {
	return c.events
}

// SendCommand sends a command to MPV
func (c *MPVIPCClient) SendCommand(cmd []interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to MPV")
	}

	cmdObj := map[string]interface{}{
		"command": cmd,
	}

	data, err := json.Marshal(cmdObj)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	data = append(data, '\n')
	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// ObserveProperty starts observing an MPV property
func (c *MPVIPCClient) ObserveProperty(id int, name string) error {
	return c.SendCommand([]interface{}{"observe_property", id, name})
}

// WaitForPlaybackStart waits for MPV to start playing the media
func (c *MPVIPCClient) WaitForPlaybackStart(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First, check if we're already playing by querying the 'idle-active' property
	if err := c.SendCommand([]interface{}{"get_property", "idle-active"}); err != nil {
		return fmt.Errorf("failed to query playback state: %w", err)
	}

	// Also observe playback-time to detect when playback actually starts
	if err := c.ObserveProperty(1, "playback-time"); err != nil {
		log.Warn("Failed to observe playback-time property", "error", err)
	}

	// Wait for either an idle-active=false response or a playback-time property change
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for MPV to start playback")
		case event, ok := <-c.events:
			if !ok {
				return fmt.Errorf("MPV connection closed while waiting for playback")
			}

			// Handle specific event types
			switch event.Event {
			case "property-change":
				// Parse the property change based on the exact format we see in the logs
				var propChange struct {
					Name string          `json:"name"`
					ID   int             `json:"id"`
					Data json.RawMessage `json:"data"`
				}

				if err := json.Unmarshal(event.Data, &propChange); err != nil {
					log.Error("Failed to parse property change", "error", err)
					continue
				}

				log.Debug("Property change parsed", "name", propChange.Name, "id", propChange.ID, "data", string(propChange.Data))

				// Check playback-time property
				if propChange.Name == "playback-time" {
					var playbackTime float64

					// Try to parse the data field as a float
					if err := json.Unmarshal(propChange.Data, &playbackTime); err != nil {
						log.Warn("Failed to parse playback-time value", "error", err)
						continue
					}

					// If playback time is positive, playback has started
					if playbackTime > 0 {
						log.Info("MPV playback has started", "time", playbackTime)
						return nil
					}
				}

				// Check idle-active property
				if propChange.Name == "idle-active" {
					var idleActive bool

					// Try to parse the data field as a boolean
					if err := json.Unmarshal(propChange.Data, &idleActive); err != nil {
						log.Warn("Failed to parse idle-active value", "error", err)
						continue
					}

					// If not idle, playback has started
					if !idleActive {
						log.Info("MPV is active (not idle)")
						return nil
					}
				}

			case "playback-restart":
				log.Info("MPV playback has started (playback-restart event)")
				return nil

			case "file-loaded":
				log.Info("MPV file has been loaded")
				return nil
			}
		}
	}
}
