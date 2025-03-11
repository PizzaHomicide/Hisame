package domain

// AnimeRepository defines the interface for anime data access
type AnimeRepository interface {
	// GetAnimeByID retrieves an anime by its ID
	GetAnimeByID(id int) (*Anime, error)

	// GetUserAnimeList retrieves a users anime list filtered by status
	GetUserAnimeList(status MediaStatus) ([]*Anime, error)

	// UpdateUserAnimeData syncs the user-specified data about an anime with AniList.  It will update everything in one
	// bulk request.
	UpdateUserAnimeData(id int, data *UserAnimeData) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// GetCurrentUser retrieves the authenticated users info
	GetCurrentUser() (*User, error)
}

// User represents an Anilist user
type User struct {
	ID     int
	Name   string
	Avatar string
}
