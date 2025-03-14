package models

// AuthCompletedMsg is sent when authentication is completed successfully
type AuthCompletedMsg struct {
	Token string
}

// AuthFailedMsg is sent when authentication fails
type AuthFailedMsg struct {
	Error string
}
