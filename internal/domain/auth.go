package domain

// AuthProvider defines the interface for authentication
type AuthProvider interface {
	// GetAuthURL returns the URL for OAuth authentication
	GetAuthURL() string

	// HandleCallback processes the OAuth callback and returns an auth token
	HandleCallback(code string) (*AuthToken, error)
}

// AuthToken represents an authentication token
type AuthToken struct {
	AccessToken string
	ExpiresIn   int
}
