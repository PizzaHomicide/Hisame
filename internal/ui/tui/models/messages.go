package models

import (
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

// AuthMsg combines auth success and failure
type AuthMsg struct {
	Success bool
	Token   string
	Error   string
}

// AnimeListMsg combines list loading success and failure
type AnimeListMsg struct {
	Success bool
	Error   error
}

// PlaybackEventType represents different playback-related events
type PlaybackEventType string

const (
	PlaybackEventEpisodeFound  PlaybackEventType = "episode_found"
	PlaybackEventSourcesLoaded PlaybackEventType = "sources_loaded"
	PlaybackEventStarted       PlaybackEventType = "started"
	PlaybackEventEnded         PlaybackEventType = "ended"
	PlaybackEventProgress      PlaybackEventType = "progress"
	PlaybackEventError         PlaybackEventType = "error"
)

// PlaybackMsg represents any playback-related event
type PlaybackMsg struct {
	Type      PlaybackEventType
	Episode   player.AllAnimeEpisodeInfo
	Sources   *player.EpisodeSourceInfo
	StreamURL string
	Progress  float64
	Error     error
}

// EpisodeEventType represents different episode-related events
type EpisodeEventType string

const (
	EpisodeEventLoaded   EpisodeEventType = "loaded"
	EpisodeEventSelected EpisodeEventType = "selected"
	EpisodeEventError    EpisodeEventType = "error"
)

// EpisodeMsg consolidates episode-related messages
type EpisodeMsg struct {
	Type     EpisodeEventType
	Episodes []player.AllAnimeEpisodeInfo
	Episode  *player.AllAnimeEpisodeInfo
	Title    string
	Error    error
}

// LoadingType represents different loading-related events
type LoadingType string

const (
	LoadingStart LoadingType = "start" // Start showing loading
	LoadingStop  LoadingType = "stop"  // Stop showing loading
)

// LoadingMsg represents a loading state change message
type LoadingMsg struct {
	Type        LoadingType
	Message     string  // Primary message to show
	Title       string  // Optional title
	ContextInfo string  // Optional context information
	ActionText  string  // Optional action text
	Operation   tea.Cmd // Optional command to run during loading
}

type AnimeListLoadResultMsg struct {
	Success   bool
	AnimeList []*domain.Anime
	Error     error
}
