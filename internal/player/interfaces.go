package player

import (
	"context"
)

// PlaybackEventType represents the type of playback event
type PlaybackEventType string

const (
	// PlaybackStarted indicates that playback has successfully started
	PlaybackStarted PlaybackEventType = "started"
	// PlaybackEnded indicates that playback has completed
	PlaybackEnded PlaybackEventType = "ended"
	// PlaybackError indicates an error during playback
	PlaybackError PlaybackEventType = "error"
	// PlaybackProgress indicates a progress update
	PlaybackProgress PlaybackEventType = "progress"
)

// PlaybackEvent represents an event from the video player
type PlaybackEvent struct {
	Type     PlaybackEventType
	Progress float64     // Percentage of progress (0-100)
	Error    error       // Error if Type is PlaybackError
	Data     interface{} // Additional data related to the event
}

// VideoPlayer defines the interface for media player implementations
type VideoPlayer interface {
	// Play starts playback of the given URL and returns a channel for playback events
	Play(ctx context.Context, url string) (<-chan PlaybackEvent, error)

	// Stop stops the current playback
	Stop() error

	// Cleanup performs any necessary cleanup
	Cleanup()
}
