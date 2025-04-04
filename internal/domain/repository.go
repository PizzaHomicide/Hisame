package domain

import "context"

// AnimeRepository defines the interface for anime data access
type AnimeRepository interface {
	// GetAllAnimeList retrieves the user's complete anime list
	GetAllAnimeList(ctx context.Context) ([]*Anime, error)

	// UpdateUserAnimeData syncs the user-specified data about an anime with AniList
	UpdateUserAnimeData(ctx context.Context, id int, data *UserAnimeData) error

	// UpdateAnime provides a structured way to update specific fields of an anime list entry
	UpdateAnime(ctx context.Context, params *AnimeUpdateParams) (*AnimeUpdateResult, error)
}

// FuzzyDate represents a date that might be incomplete (missing day or month)
type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// AnimeUpdateParams defines the parameters that can be updated for an anime list entry
type AnimeUpdateParams struct {
	MediaID     int        `json:"mediaId"` // Required - The ID of the anime to update
	Status      string     `json:"status,omitempty"`
	Progress    *int       `json:"progress,omitempty"`
	Score       *float64   `json:"score,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	StartedAt   *FuzzyDate `json:"startedAt,omitempty"`
	CompletedAt *FuzzyDate `json:"completedAt,omitempty"`
}

// AnimeUpdateResult contains information about the result of an anime update operation
type AnimeUpdateResult struct {
	EntryID        int         // The list entry ID
	MediaID        int         // The media ID that was updated
	Status         MediaStatus // The status after the update
	Progress       int         // The progress after the update
	Score          float64     // The score after the update
	Notes          string      // The notes after the update
	UpdatedAt      int         // The timestamp when the update occurred
	StartDate      string      // The start date after the update
	CompletionDate string      // The completion date after the update
}

// ToAnimeUpdateVariables converts the update params to a variables map for GraphQL
func (p *AnimeUpdateParams) ToAnimeUpdateVariables() map[string]interface{} {
	variables := map[string]interface{}{
		"mediaId": p.MediaID,
	}

	if p.Status != "" {
		variables["status"] = p.Status
	}

	if p.Progress != nil {
		variables["progress"] = *p.Progress
	}

	if p.Score != nil {
		variables["score"] = *p.Score
	}

	if p.Notes != nil {
		variables["notes"] = *p.Notes
	}

	if p.StartedAt != nil {
		// Only include non-zero date components
		startedAtMap := map[string]int{}
		if p.StartedAt.Year > 0 {
			startedAtMap["year"] = p.StartedAt.Year
		}
		if p.StartedAt.Month > 0 {
			startedAtMap["month"] = p.StartedAt.Month
		}
		if p.StartedAt.Day > 0 {
			startedAtMap["day"] = p.StartedAt.Day
		}

		// Only add if we have at least one date component
		if len(startedAtMap) > 0 {
			variables["startedAt"] = startedAtMap
		}
	}

	if p.CompletedAt != nil {
		// Only include non-zero date components
		completedAtMap := map[string]int{}
		if p.CompletedAt.Year > 0 {
			completedAtMap["year"] = p.CompletedAt.Year
		}
		if p.CompletedAt.Month > 0 {
			completedAtMap["month"] = p.CompletedAt.Month
		}
		if p.CompletedAt.Day > 0 {
			completedAtMap["day"] = p.CompletedAt.Day
		}

		// Only add if we have at least one date component
		if len(completedAtMap) > 0 {
			variables["completedAt"] = completedAtMap
		}
	}

	return variables
}
