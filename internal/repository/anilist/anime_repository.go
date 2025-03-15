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
							Romaji  string
							English string
							Native  string
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
					Romaji:  entry.Media.Title.Romaji,
					English: entry.Media.Title.English,
					Native:  entry.Media.Title.Native,
				},
				CoverImage:   entry.Media.CoverImage.Large,
				Episodes:     entry.Media.Episodes,
				Status:       entry.Media.Status,
				Format:       entry.Media.Format,
				Season:       entry.Media.Season,
				SeasonYear:   fmt.Sprintf("%d", entry.Media.SeasonYear),
				AverageScore: entry.Media.AverageScore,
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
	panic("Not yet implemented")
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
