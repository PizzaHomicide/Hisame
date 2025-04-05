package anilist

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
)

type AnimeRepository struct {
	client *Client
}

func NewAnimeRepository(client *Client) domain.AnimeRepository {
	return &AnimeRepository{
		client: client,
	}
}

func (r *AnimeRepository) GetAllAnimeList(ctx context.Context) ([]*domain.Anime, error) {
	query := `
        query ($userId: Int) {
            MediaListCollection(userId: $userId, type: ANIME) {
                lists {
                    entries {
                        media {
                            id
                            title {
                                romaji
                                english
                                native
								userPreferred
                            }
                            coverImage {
                                large
                            }
                            episodes
                            nextAiringEpisode {
                                episode
                                airingAt
                                timeUntilAiring
                            }
                            status
                            format
                            season
                            seasonYear
                            averageScore
							synonyms
                        }
                        status
                        score
                        progress
                        startedAt { year month day }
                        completedAt { year month day }
                        notes
                    }
                }
            }
        }
    `

	variables := map[string]interface{}{
		"userId": r.client.user.ID,
	}

	var response struct {
		MediaListCollection struct {
			Lists []struct {
				Entries []struct {
					Media struct {
						ID    int
						Title struct {
							Romaji        string
							English       string
							Native        string
							UserPreferred string
						}
						CoverImage struct {
							Large string
						}
						Episodes          int
						NextAiringEpisode *struct {
							Episode         int
							AiringAt        int64
							TimeUntilAiring int64
						}
						Status       string
						Format       string
						Season       string
						SeasonYear   int
						AverageScore float64
						Synonyms     []string
					}
					Status    string
					Score     float64
					Progress  int
					StartedAt struct {
						Year  int
						Month int
						Day   int
					}
					CompletedAt struct {
						Year  int
						Month int
						Day   int
					}
					Notes string
				}
			}
		}
	}

	if err := r.client.Query(ctx, query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch anime list: %w", err)
	}

	var animeList []*domain.Anime

	for _, list := range response.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			anime := &domain.Anime{
				ID: entry.Media.ID,
				Title: domain.AnimeTitle{
					Romaji:    entry.Media.Title.Romaji,
					English:   entry.Media.Title.English,
					Native:    entry.Media.Title.Native,
					Preferred: entry.Media.Title.UserPreferred,
				},
				CoverImage:   entry.Media.CoverImage.Large,
				Episodes:     entry.Media.Episodes,
				Status:       entry.Media.Status,
				Format:       entry.Media.Format,
				Season:       entry.Media.Season,
				SeasonYear:   fmt.Sprintf("%d", entry.Media.SeasonYear),
				AverageScore: entry.Media.AverageScore,
				Synonyms:     entry.Media.Synonyms,
				UserData: &domain.UserAnimeData{
					Status:    domain.MediaStatus(entry.Status),
					Score:     entry.Score,
					Progress:  entry.Progress,
					StartDate: formatDate(entry.StartedAt.Year, entry.StartedAt.Month, entry.StartedAt.Day),
					EndDate:   formatDate(entry.CompletedAt.Year, entry.CompletedAt.Month, entry.CompletedAt.Day),
					Notes:     entry.Notes,
				},
			}

			if entry.Media.NextAiringEpisode != nil {
				anime.NextAiringEp = &domain.AiringSchedule{
					Episode:      entry.Media.NextAiringEpisode.Episode,
					AiringAt:     entry.Media.NextAiringEpisode.AiringAt,
					TimeUntilAir: entry.Media.NextAiringEpisode.TimeUntilAiring,
				}
			}

			animeList = append(animeList, anime)
		}
	}

	log.Info("Fetched complete anime list", "count", len(animeList))
	return animeList, nil
}

func (r *AnimeRepository) UpdateUserAnimeData(ctx context.Context, id int, data *domain.UserAnimeData) error {
	mutation := `
		mutation ($mediaId: Int, $status: MediaListStatus, $score: Float, $progress: Int, $notes: String) {
			SaveMediaListEntry(
				mediaId: $mediaId, 
				status: $status, 
				score: $score, 
				progress: $progress,
				notes: $notes
			) {
				id
				status
				score
				progress
				notes
			}
		}
	`

	// Convert domain.MediaStatus to string for the GraphQL API
	variables := map[string]interface{}{
		"mediaId":  id,
		"status":   string(data.Status),
		"score":    data.Score,
		"progress": data.Progress,
		"notes":    data.Notes,
	}

	// For date fields, we would need to parse and convert the format
	// This can be added if needed for start/end dates

	log.Debug("Updating anime data",
		"mediaId", id,
		"status", data.Status,
		"score", data.Score,
		"progress", data.Progress)

	var response struct {
		SaveMediaListEntry struct {
			ID       int     `json:"id"`
			Status   string  `json:"status"`
			Score    float64 `json:"score"`
			Progress int     `json:"progress"`
			Notes    string  `json:"notes"`
		}
	}

	if err := r.client.Query(ctx, mutation, variables, &response); err != nil {
		log.Error("Failed to update anime data", "error", err, "mediaId", id)
		return fmt.Errorf("failed to update anime data: %w", err)
	}

	log.Info("Successfully updated anime data",
		"mediaId", id,
		"listEntryId", response.SaveMediaListEntry.ID,
		"status", response.SaveMediaListEntry.Status,
		"progress", response.SaveMediaListEntry.Progress)

	return nil
}

// UpdateAnime provides a structured way to update specific fields of an anime list entry
func (r *AnimeRepository) UpdateAnime(ctx context.Context, params *domain.AnimeUpdateParams) (*domain.AnimeUpdateResult, error) {
	mutation := `
		mutation (
			$mediaId: Int, 
			$status: MediaListStatus, 
			$score: Float, 
			$progress: Int, 
			$notes: String,
			$startedAt: FuzzyDateInput,
			$completedAt: FuzzyDateInput
		) {
			SaveMediaListEntry(
				mediaId: $mediaId, 
				status: $status, 
				score: $score, 
				progress: $progress,
				notes: $notes,
				startedAt: $startedAt,
				completedAt: $completedAt
			) {
				id
				mediaId
				status
				score
				progress
				notes
				updatedAt
				startedAt {
					year
					month
					day
				}
				completedAt {
					year
					month
					day
				}
			}
		}
	`

	// Convert params to variables map
	variables := params.ToAnimeUpdateVariables()

	log.Debug("Updating anime data",
		"mediaId", params.MediaID,
		"variables", variables)

	var response struct {
		SaveMediaListEntry struct {
			ID        int     `json:"id"`
			MediaID   int     `json:"mediaId"`
			Status    string  `json:"status"`
			Score     float64 `json:"score"`
			Progress  int     `json:"progress"`
			Notes     string  `json:"notes"`
			UpdatedAt int     `json:"updatedAt"`
			StartedAt struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"startedAt"`
			CompletedAt struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"completedAt"`
		}
	}

	if err := r.client.Query(ctx, mutation, variables, &response); err != nil {
		log.Error("Failed to update anime data", "error", err, "mediaId", params.MediaID)
		return nil, fmt.Errorf("failed to update anime data: %w", err)
	}

	// Create the result
	result := &domain.AnimeUpdateResult{
		EntryID:   response.SaveMediaListEntry.ID,
		MediaID:   response.SaveMediaListEntry.MediaID,
		Status:    domain.MediaStatus(response.SaveMediaListEntry.Status),
		Progress:  response.SaveMediaListEntry.Progress,
		Score:     response.SaveMediaListEntry.Score,
		Notes:     response.SaveMediaListEntry.Notes,
		UpdatedAt: response.SaveMediaListEntry.UpdatedAt,
	}

	// Check for start date
	startedAt := response.SaveMediaListEntry.StartedAt
	if startedAt.Year > 0 {
		result.StartDate = formatDate(startedAt.Year, startedAt.Month, startedAt.Day)
	}

	// Check for completion date
	completedAt := response.SaveMediaListEntry.CompletedAt
	if completedAt.Year > 0 {
		result.CompletionDate = formatDate(completedAt.Year, completedAt.Month, completedAt.Day)
	}

	log.Info("Successfully updated anime data",
		"mediaId", result.MediaID,
		"listEntryId", result.EntryID,
		"status", result.Status,
		"progress", result.Progress,
		"updatedAt", result.UpdatedAt)

	return result, nil
}

func formatDate(year, month, day int) string {
	if year == 0 {
		return ""
	}

	if month == 0 {
		return fmt.Sprintf("%d", year)
	}

	if day == 0 {
		return fmt.Sprintf("%d-%02d", year, month)
	}

	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}
