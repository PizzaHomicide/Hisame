package domain

import "context"

// AnimeRepository defines the interface for anime data access
type AnimeRepository interface {
	// GetAllAnimeList retrieves the user's complete anime list
	GetAllAnimeList(ctx context.Context) ([]*Anime, error)

	// UpdateUserAnimeData syncs the user-specified data about an anime with AniList
	UpdateUserAnimeData(ctx context.Context, id int, data *UserAnimeData) error
}
