package models

// View represents a specific UI view in the application
type View string

// Available views in the application
const (
	ViewAuth         View = "auth"
	ViewAnimeList    View = "anime-list"
	ViewAnimeDetails View = "anime-detail"
	ViewStatus       View = "status"
)

// Modal represents a UI intended to be temporarily shown to the user before returning to the original view
type Modal string

// Available modals in the application
const (
	ModalNone          Modal = "none"
	ModalHelp          Modal = "help"
	ModalEpisodeSelect Modal = "episode_select"
)
