package player

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/log"
	"net/http"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
)

const (
	allAnimeGraphQLURL = "https://api.allanime.day/api"
	allAnimeUserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

// AllAnimeClient is responsible for communicating with the AllAnime API
type AllAnimeClient struct {
	client *graphql.Client
}

// NewAllAnimeClient creates a new AllAnime client
func NewAllAnimeClient() *AllAnimeClient {
	// Create a custom HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create a new GraphQL client with the custom HTTP client
	client := graphql.NewClient(allAnimeGraphQLURL, graphql.WithHTTPClient(httpClient))

	return &AllAnimeClient{
		client: client,
	}
}

// AiredDate represents a date in the AllAnime API
type AiredDate struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Date   int `json:"date"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// ToTime converts the AiredDate to a time.Time
func (a AiredDate) ToTime() time.Time {
	return time.Date(a.Year, time.Month(a.Month), a.Date, a.Hour, a.Minute, 0, 0, time.UTC)
}

// Season represents a season in the AllAnime API
type Season struct {
	Quarter string `json:"quarter"`
	Year    int    `json:"year"`
}

// AllAnimeShow represents a show in the AllAnime API
type AllAnimeShow struct {
	ID                      string    `json:"_id"`
	Name                    string    `json:"name"`
	EnglishName             string    `json:"englishName"`
	NativeName              string    `json:"nativeName"`
	TrustedAltNames         []string  `json:"trustedAltNames"`
	AniListID               string    `json:"aniListId"`
	Season                  Season    `json:"season"`
	AiredStart              AiredDate `json:"airedStart"`
	AiredEnd                AiredDate `json:"airedEnd"`
	AvailableEpisodesDetail struct {
		Sub []string `json:"sub"`
		Dub []string `json:"dub"`
	} `json:"availableEpisodesDetail"`
}

// GetAniListID returns the AniListID as an integer
func (s AllAnimeShow) GetAniListID() int {
	if s.AniListID == "" || s.AniListID == "null" {
		return 0
	}
	id, err := strconv.Atoi(s.AniListID)
	if err != nil {
		// Note that this is actually expected in some edge cases (like allanime splitting a season into halves/cours), but it's uncommon
		log.Warn("Failed to convert AniListID to int", "id", s.AniListID, "allanime_id", s.ID, "title", s.EnglishName)
		return 0
	}
	return id
}

// GetAvailableEpisodes returns the available episodes for the given translation type
func (s AllAnimeShow) GetAvailableEpisodes(translationType string) []string {
	switch translationType {
	case "sub":
		return s.AvailableEpisodesDetail.Sub
	case "dub":
		return s.AvailableEpisodesDetail.Dub
	default:
		return s.AvailableEpisodesDetail.Sub // Default to sub
	}
}

// ShowSearchResponse represents the response from the shows search GraphQL query
type ShowSearchResponse struct {
	Shows struct {
		Edges []AllAnimeShow `json:"edges"`
	} `json:"shows"`
}

// SearchShows searches for shows matching the given query
func (c *AllAnimeClient) SearchShows(ctx context.Context, query string, translationType string) ([]AllAnimeShow, error) {
	// Create the GraphQL request
	req := graphql.NewRequest(`
		query ($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
			shows(
				search: $search
				limit: $limit
				page: $page
				translationType: $translationType
				countryOrigin: $countryOrigin
			) {
				edges {
					_id
					name
					englishName
					nativeName
					trustedAltNames
					availableEpisodesDetail
					season
					airedStart
					airedEnd
					aniListId
				}
			}
		}
	`)

	// Set the variables
	req.Var("search", map[string]interface{}{
		"allowAdult":   true,
		"allowUnknown": false,
		"query":        query,
	})
	req.Var("limit", 20)
	// TODO:  Paging support.  But 20 is probably safe for the specific queries we're running.  Will support paging if I ever find a case where things don't work.
	req.Var("page", 1)
	req.Var("translationType", translationType)
	req.Var("countryOrigin", "ALL")

	// Set the user agent header
	req.Header.Set("User-Agent", allAnimeUserAgent)

	log.Debug("Before request")
	// Execute the request
	var response ShowSearchResponse
	if err := c.client.Run(ctx, req, &response); err != nil {
		log.Debug("Error executing request", "err", err)
		return nil, fmt.Errorf("error searching shows: %w", err)
	}

	log.Debug("Search shows", "response", response)

	return response.Shows.Edges, nil
}

// EpisodeSource represents a single streaming source for an episode
type EpisodeSource struct {
	SourceURL  string  `json:"sourceUrl"`
	Priority   float64 `json:"priority"`
	SourceName string  `json:"sourceName"`
	Type       string  `json:"type"` // "iframe", "player", etc.
	ClassName  string  `json:"className"`
	StreamerID string  `json:"streamerId"`
	Downloads  *struct {
		SourceName  string `json:"sourceName"`
		DownloadURL string `json:"downloadUrl"`
	} `json:"downloads,omitempty"`
	Sandbox string `json:"sandbox,omitempty"`
}

// EpisodeSourceResponse represents the structure of the GraphQL response
type EpisodeSourceResponse struct {
	Episode struct {
		EpisodeString string          `json:"episodeString"`
		SourceUrls    []EpisodeSource `json:"sourceUrls"`
	} `json:"episode"`
}

// GetEpisodeSources fetches the available streaming sources for a specific episode
func (c *AllAnimeClient) GetEpisodeSources(ctx context.Context, showID string, episodeNum string, translationType string) ([]EpisodeSource, error) {
	// Create the GraphQL request
	req := graphql.NewRequest(`
		query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) {
			episode(
				showId: $showId
				translationType: $translationType
				episodeString: $episodeString
			) {
				episodeString
				sourceUrls
			}
		}
	`)

	// Set the variables
	req.Var("showId", showID)
	req.Var("translationType", translationType)
	req.Var("episodeString", episodeNum)

	// Set the user agent header
	req.Header.Set("User-Agent", allAnimeUserAgent)

	log.Debug("Fetching episode sources", "showId", showID, "episodeNum", episodeNum, "translationType", translationType)

	// Execute the request
	var response EpisodeSourceResponse
	if err := c.client.Run(ctx, req, &response); err != nil {
		log.Error("Error fetching episode sources", "error", err)
		return nil, fmt.Errorf("error fetching episode sources: %w", err)
	}

	sources := response.Episode.SourceUrls
	log.Debug("Episode sources retrieved successfully", "count", len(sources))
	return sources, nil
}
