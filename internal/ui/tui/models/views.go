package models

import tea "github.com/charmbracelet/bubbletea"

// View represents a specific UI view in the application
type View string

// Available views in the application
const (
	ViewAuth          View = "auth"
	ViewAnimeList     View = "anime-list"
	ViewHelp          View = "help"
	ViewEpisodeSelect View = "episode-select"
	ViewLoading       View = "loading"
	ViewAnimeDetails  View = "anime-details"
	ViewMenu          View = "menu"
)

// Model is the interface that all our models should implement
type Model interface {
	// Init initializes the model and returns any initial command
	Init() tea.Cmd

	// Update handles messages and returns the updated model and any command
	Update(msg tea.Msg) (Model, tea.Cmd)

	// View renders the model to a string
	View() string

	// Resize updates a models width & height
	Resize(width, height int)

	// ViewType returns the type of the view
	ViewType() View
}
