package anilist

import (
	"context"
	"errors"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/machinebox/graphql"
	"net/url"
	"strings"
	"time"
)

// Client is the generic AniList client for making queries to the AniList graphql API
type Client struct {
	client    *graphql.Client
	authToken string
	user      domain.User
}

func (c *Client) GetUser() domain.User {
	return c.user
}

func NewClient(authToken string) (*Client, error) {
	if authToken == "" {
		log.Error("AniList Client authToken is empty.")
		return nil, fmt.Errorf("AniList Client authToken is empty")
	}

	client := graphql.NewClient("https://graphql.anilist.co")
	c := &Client{
		client:    client,
		authToken: authToken,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user, err := c.fetchUserProfile(ctx)
	if err != nil {
		return nil, err
	}

	c.user = *user
	return c, nil
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	req := graphql.NewRequest(query)

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	for key, value := range variables {
		req.Var(key, value)
	}

	return c.client.Run(ctx, req, result)
}

type NetworkError struct {
	Err error
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e NetworkError) Unwrap() error {
	return e.Err
}

// Update the fetchUserProfile method to detect network errors
func (c *Client) fetchUserProfile(ctx context.Context) (*domain.User, error) {
	query := `
        query {
            Viewer {
                id
                name
                avatar {
                    medium
                }
                siteUrl
                statistics {
                    anime {
                        count
                        episodesWatched
                    }
                    manga {
                        count
                        chaptersRead
                    }
                }
                options {
                    titleLanguage
                    displayAdultContent
                }
            }
        }
    `

	var response struct {
		Viewer struct {
			ID         int
			Name       string
			Avatar     struct{ Medium string }
			SiteUrl    string
			Statistics struct {
				Anime struct {
					Count           int
					EpisodesWatched int `json:"episodesWatched"`
				}
				Manga struct {
					Count        int
					ChaptersRead int `json:"chaptersRead"`
				}
			}
		}
	}

	if err := c.Query(ctx, query, nil, &response); err != nil {
		// Check if this is a network error
		var netErr *url.Error
		if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary() ||
			strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "no such host") ||
			strings.Contains(err.Error(), "i/o timeout")) {
			return nil, NetworkError{Err: err}
		}
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	if response.Viewer.ID == 0 {
		return nil, fmt.Errorf("invalid or unauthorized token")
	}

	log.Info("Fetched user profile", "id", response.Viewer.ID)

	return &domain.User{
		ID:      response.Viewer.ID,
		Name:    response.Viewer.Name,
		Avatar:  response.Viewer.Avatar.Medium,
		SiteURL: response.Viewer.SiteUrl,
		Statistics: domain.UserStatistics{
			AnimeCount:      response.Viewer.Statistics.Anime.Count,
			MangaCount:      response.Viewer.Statistics.Manga.Count,
			EpisodesWatched: response.Viewer.Statistics.Anime.EpisodesWatched,
			ChaptersRead:    response.Viewer.Statistics.Manga.ChaptersRead,
		},
	}, nil
}
