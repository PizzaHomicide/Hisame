package models

import (
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/player"
	"github.com/PizzaHomicide/hisame/internal/repository/anilist"
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
	Anime     *domain.Anime
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

// TokenValidationMsg represents the result of validating an authentication token
type TokenValidationMsg struct {
	Valid     bool            // Whether the token is valid
	Client    *anilist.Client // The initialized client if token is valid
	User      *domain.User    // User information if token is valid
	Error     error           // Error that occurred during validation, if any
	IsNetwork bool            // Whether the error was a network-related error
}

// AnimeUpdatedMsg indicates an anime in the list has been updated
type AnimeUpdatedMsg struct {
	Success bool
	AnimeID int
	Message string
	Error   error
}

// PlaybackCompletedMsg is used to transmit playback completion from goroutines
type PlaybackCompletedMsg struct {
	AnimeID       int
	EpisodeNumber int
	Progress      float64
}

// AnimeDetailsMsg is sent when a user wants to view the details for an anime
type AnimeDetailsMsg struct {
	Anime *domain.Anime
}

// HandledMsg is used when a model wants to bubble up the fact that it handled a message
// and that further processing is likely not required (though the orchestration layer still
// CAN do further processing if it deems necessary).
type HandledMsg struct {
	Message string // A message that will be debug logged
}

func Handled(message string) tea.Cmd {
	return func() tea.Msg {
		return HandledMsg{Message: message}
	}
}

// ShowMenuMsg is sent when a menu should be displayed
type ShowMenuMsg struct {
	Menu *MenuModel
}

// MenuSelectionMsg is sent when a menu item is selected
type MenuSelectionMsg struct {
	CloseMenu bool    // Whether to close the menu after selection
	NextMsg   tea.Msg // The message to propagate next
}

// PlayNextEpisodeMsg is sent when the next episode of a given anime should be played
// Thoughts:  Consider if this should be a more populated message.  Right now it expects the anime list model to handle
//
//	           it, but what if we wanted something else to?  We could provide all the info required here, then some playback
//				  handler could deal with it.
type PlayNextEpisodeMsg struct {
	AnimeID int
}
