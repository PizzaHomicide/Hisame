package models

// AuthCompletedMsg is sent when authentication is completed successfully
type AuthCompletedMsg struct {
	Token string
}

// AuthFailedMsg is sent when authentication fails
type AuthFailedMsg struct {
	Error string
}

// AnimeListLoadedMsg is sent when the anime list is loaded
type AnimeListLoadedMsg struct{}

// AnimeListErrorMsg is sent when there's an error loading the anime list
type AnimeListErrorMsg struct {
	Error error
}
