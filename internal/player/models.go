package player

import (
	"context"
	"time"
)

// PlayerType defines the type of media player to use
type PlayerType string

const (
	// PlayerTypeMPV represents the MPV player
	PlayerTypeMPV PlayerType = "mpv"
	// PlayerTypeCustom represents a custom player executable
	PlayerTypeCustom PlayerType = "custom"
)

// AllAnimeEpisodeInfo contains information about an available episode
type AllAnimeEpisodeInfo struct {
	// The ID of the anime on allanime
	AllAnimeID string
	// The overall episode number (adjusted for multi-season shows)
	OverallEpisodeNumber int
	// The episode number as represented on allanime
	AllAnimeEpisodeNumber string
	// The title of the anime on allanime
	AllAnimeName string
	// Additional titles from allanime
	PreferredTitle string
	// The alt names of the show
	AltNames []string
	// Airing date if available
	AirDate time.Time
	// The AniList ID if available
	AniListID int
	// The season information
	Season string
	Year   int
	// Whether this was matched by AniList ID or by synonyms
	MatchType string
}

// FindEpisodesResult contains the complete result of finding episodes
type FindEpisodesResult struct {
	// The list of episodes found
	Episodes []AllAnimeEpisodeInfo
	// The raw AllAnime show data
	RawShows []AllAnimeShow
}

// Service defines the interface for the player service
type Service interface {
	// FindEpisodes finds all available episodes for an anime
	FindEpisodes(ctx context.Context, animeID int, title string, synonyms []string) (*FindEpisodesResult, error)
}
