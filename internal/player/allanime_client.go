package player

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

	// Execute the request
	var response ShowSearchResponse
	if err := c.client.Run(ctx, req, &response); err != nil {
		log.Debug("Error executing request", "err", err)
		return nil, fmt.Errorf("error searching shows: %w", err)
	}

	log.Debug("Search shows", "response", response, "query", query)

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
	var response map[string]interface{}
	if err := c.client.Run(ctx, req, &response); err != nil {
		log.Error("Error fetching episode sources", "error", err)
		return nil, fmt.Errorf("error fetching episode sources: %w", err)
	}

	// Check if the response contains a tobeparsed field (encrypted response)
	if tobeparsed, ok := response["tobeparsed"].(string); ok && tobeparsed != "" {
		// Decrypt the tobeparsed data
		decrypted, err := c.decryptTobeparsed(tobeparsed)
		if err != nil {
			return nil, fmt.Errorf("error decrypting tobeparsed data: %w", err)
		}

		// Parse the decrypted JSON to get sources
		var decryptedResponse EpisodeSourceResponse
		if err := json.Unmarshal([]byte(decrypted), &decryptedResponse); err != nil {
			return nil, fmt.Errorf("error parsing decrypted response: %w", err)
		}

		sources := decryptedResponse.Episode.SourceUrls
		log.Debug("Episode sources retrieved successfully (decrypted)", "count", len(sources))
		return sources, nil
	}

	// Process normal response
	var normalResponse EpisodeSourceResponse
	if err := mapToStruct(response, &normalResponse); err != nil {
		return nil, fmt.Errorf("error parsing normal response: %w", err)
	}

	sources := normalResponse.Episode.SourceUrls
	log.Debug("Episode sources retrieved successfully", "count", len(sources))
	return sources, nil
}

// decryptTobeparsed decrypts the AES-256-CTR encrypted tobeparsed field
// This replicates the OpenSSL decryption logic from ani-cli:
// 1. Key = SHA256("SimtVuagFbGR2K7P") as hex string, then decoded
// 2. IV = 12 bytes from data + "00000002" as hex, then decoded to 16 bytes
func (c *AllAnimeClient) decryptTobeparsed(tobeparsed string) (string, error) {
	// Decode base64
	encryptedData, err := base64.StdEncoding.DecodeString(tobeparsed)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode: %w", err)
	}

	// Extract IV (first 12 bytes) and ciphertext
	// The last 16 bytes appear to be padding/auth tag that we should skip
	if len(encryptedData) < 28 { // 12 IV + at least 1 ciphertext + 15 padding minimum
		return "", fmt.Errorf("encrypted data too short: %d bytes", len(encryptedData))
	}

	// Extract 12-byte IV
	ivBytes := encryptedData[:12]
	// Ciphertext is everything between IV and the last 16 bytes
	ciphertext := encryptedData[12 : len(encryptedData)-16]

	// Generate key: SHA256(secret) as hex string, then decode to bytes
	// This matches: printf '%s' 'SimtVuagFbGR2K7P' | openssl dgst -sha256 -binary | od -A n -t x1 | tr -d ' \n'
	keyHash := sha256.Sum256([]byte("SimtVuagFbGR2K7P"))
	keyHex := hex.EncodeToString(keyHash[:]) // 64-character hex string
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode key hex: %w", err)
	}

	// Build the counter IV: IV bytes as hex + "00000002", then decode to 16 bytes
	// This matches: ctr="${iv}00000002" in ani-cli
	ivHex := hex.EncodeToString(ivBytes) // 24-character hex string (12 bytes)
	ctrHex := ivHex + "00000002"         // Append 8 hex characters (4 bytes: 0x00,0x00,0x00,0x02)
	ctr, err := hex.DecodeString(ctrHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode counter hex: %w", err)
	}

	// Create AES cipher with the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create CTR mode stream with the counter
	stream := cipher.NewCTR(block, ctr)

	// Decrypt
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}

// mapToStruct converts a map to a struct using JSON marshaling/unmarshaling
func mapToStruct(m map[string]interface{}, out interface{}) error {
	tmp, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(tmp, out)
}
