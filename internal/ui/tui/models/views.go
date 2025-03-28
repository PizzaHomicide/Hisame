package models

// View represents a specific UI view in the application
type View string

// Available views in the application
const (
	ViewAuth          View = "auth"
	ViewAnimeList     View = "anime-list"
	ViewHelp          View = "help"
	ViewEpisodeSelect View = "episode-select"
	ViewLoading       View = "loading"
)
