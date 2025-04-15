package player

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// MatchTypeAniList indicates the show was matched by AniList ID
	MatchTypeAniList = "anilist"
	// MatchTypeSynonym indicates the show was matched by synonym
	MatchTypeSynonym = "synonym"
)

// PlayerService implements the Service interface
type PlayerService struct {
	config      *config.Config
	animeClient *AllAnimeClient
}

// NewPlayerService creates a new player service
func NewPlayerService(config *config.Config) *PlayerService {
	return &PlayerService{
		config:      config,
		animeClient: NewAllAnimeClient(),
	}
}

// FindEpisodes implements the Service FindEpisodes method
func (s *PlayerService) FindEpisodes(ctx context.Context, animeID int, title *domain.AnimeTitle, synonyms []string) (*FindEpisodesResult, error) {
	log.Debug("Finding episodes", "title", title.Preferred, "id", animeID, "synonyms", synonyms)

	// Search for shows matching the anime title.  Cycles through each language looking for a match, as sometimes
	// we find one for one language, but not another.
	titles := []string{title.Native, title.English, title.Romaji}
	var allShows []AllAnimeShow

	// Try each title format
	for _, title := range titles {
		if title == "" {
			continue // Skip empty titles
		}

		shows, err := s.animeClient.SearchShows(ctx, title, s.config.Player.TranslationType)
		if err != nil {
			log.Warn("Error searching with title format", "title", title, "error", err)
			continue // Try next format on error
		}

		// Add these shows to our collection
		allShows = append(allShows, shows...)
	}
	// Deduplicate by AllAnime ID
	shows := deduplicateShows(allShows)

	if len(shows) == 0 {
		return nil, errors.New("no candidate shows found")
	}

	log.Debug("Found candidate shows on allanime", "count", len(shows))

	// Find all matching shows (either by AniList ID or by synonyms)
	var matchedShows []AllAnimeShow

	for _, show := range shows {
		aniListID := show.GetAniListID()

		if aniListID == animeID {
			// Direct match by AniList ID
			log.Debug("Found direct AniList ID match", "allanime_id", show.ID, "name", show.Name, "anilist_id", aniListID)
			matchedShows = append(matchedShows, show)
		} else if aniListID == 0 && s.matchesByTitleOrSynonyms(title, synonyms, show) {
			// Match by title or synonyms for shows without AniList ID
			log.Debug("Found match by title or synonym", "allanime_id", show.ID, "name", show.Name)
			matchedShows = append(matchedShows, show)
		}
	}

	if len(matchedShows) == 0 {
		return nil, errors.New("no matching shows found after filtering")
	}

	// Sort matched shows chronologically by air date
	sort.Slice(matchedShows, func(i, j int) bool {
		// If both have valid air dates, compare them
		startI := matchedShows[i].AiredStart
		startJ := matchedShows[j].AiredStart

		// Check if we have valid years
		// TODO:  Check over this before committing..  Feels a bit redundant..  We probably always have correct years.
		if startI.Year > 0 && startJ.Year > 0 {
			timeI := startI.ToTime()
			timeJ := startJ.ToTime()
			return timeI.Before(timeJ)
		}

		// Fallback to ID comparison if we don't have valid dates
		return matchedShows[i].ID < matchedShows[j].ID
	})

	// Build the episode list from matched shows
	result := s.buildEpisodeList(matchedShows, animeID, title)

	log.Debug("Built episode list", "matched_show_count", len(matchedShows), "episode_count", len(result.Episodes), "title", title)

	return result, nil
}

func deduplicateShows(shows []AllAnimeShow) []AllAnimeShow {
	seen := make(map[string]bool)
	var result []AllAnimeShow

	for _, show := range shows {
		if !seen[show.ID] {
			seen[show.ID] = true
			result = append(result, show)
		}
	}

	return result
}

// matchesByTitleOrSynonyms checks if a show matches the anime by title or synonyms
func (s *PlayerService) matchesByTitleOrSynonyms(title *domain.AnimeTitle, synonyms []string, show AllAnimeShow) bool {
	// Check if the anime title matches any of the show's names
	if strings.ToLower(show.Name) == strings.ToLower(title.Romaji) ||
		strings.ToLower(show.EnglishName) == strings.ToLower(title.English) ||
		strings.ToLower(show.NativeName) == strings.ToLower(title.Native) {
		log.Debug("AllAnimeName match found", "title", title, "allanime_name", show.Name,
			"allanime_englishname", show.EnglishName, "allanime_nativename", show.NativeName)
		return true
	}

	// Check if any of the show's alt names match any of the anime's synonyms
	for _, altName := range show.TrustedAltNames {
		altNameLower := strings.ToLower(altName)

		// Check against anime synonyms
		for _, synonym := range synonyms {
			if altNameLower == strings.ToLower(synonym) {
				log.Debug("Synonym + alt name match found", "synonym", synonym, "title", title, "alt_name", altName)
				return true
			}
		}
	}

	// No matches found
	return false
}

// buildEpisodeList builds a chronologically ordered list of episodes from the matched shows
func (s *PlayerService) buildEpisodeList(shows []AllAnimeShow, animeID int, titles *domain.AnimeTitle) *FindEpisodesResult {
	var episodes []AllAnimeEpisodeInfo
	episodeOffset := 0

	// Process each show in chronological order
	for _, show := range shows {
		availableEps := show.GetAvailableEpisodes(s.config.Player.TranslationType)

		// Skip shows with no available episodes
		if len(availableEps) == 0 {
			continue
		}

		// Convert episode strings to numbers and sort
		var episodeNums []int
		episodeMap := make(map[int]string)
		for _, ep := range availableEps {
			epNum, err := strconv.Atoi(ep)
			if err != nil {
				log.Warn("Could not parse episode number", "episode", ep, "error", err)
				continue
			}
			episodeNums = append(episodeNums, epNum)
			episodeMap[epNum] = ep
		}
		sort.Ints(episodeNums)

		// Determine match type
		matchType := MatchTypeSynonym
		if show.GetAniListID() == animeID && animeID != 0 {
			matchType = MatchTypeAniList
		}

		// Create episode info for each episode
		for _, epNum := range episodeNums {
			epStr := episodeMap[epNum]

			// Calculate overall episode number
			overallEpNum := epNum + episodeOffset

			episodes = append(episodes, AllAnimeEpisodeInfo{
				AllAnimeID:            show.ID,
				OverallEpisodeNumber:  overallEpNum,
				AllAnimeEpisodeNumber: epStr,
				AllAnimeName:          show.Name,
				PreferredTitle:        titles.Preferred,
				AltNames:              show.TrustedAltNames,
				AirDate:               show.AiredStart.ToTime(),
				AniListID:             show.GetAniListID(),
				Season:                show.Season.Quarter,
				Year:                  show.Season.Year,
				MatchType:             matchType,
			})
		}

		// Update the offset for the next show
		if len(episodeNums) > 0 {
			maxEpNum := episodeNums[len(episodeNums)-1]
			episodeOffset += maxEpNum
		}
	}

	return &FindEpisodesResult{
		Episodes: episodes,
		RawShows: shows,
	}
}

// EpisodeSourceInfo contains information about available sources for an episode
type EpisodeSourceInfo struct {
	AnimeName       string
	EpisodeNumber   string
	AllAnimeID      string
	Sources         []EpisodeSource
	TranslationType string
}

// GetEpisodeSources fetches all available sources for a specific episode and filters to supported types
func (s *PlayerService) GetEpisodeSources(ctx context.Context, animeInfo AllAnimeEpisodeInfo) (*EpisodeSourceInfo, error) {
	log.Debug("Getting episode sources",
		"allAnimeID", animeInfo.AllAnimeID,
		"episodeNumber", animeInfo.AllAnimeEpisodeNumber,
		"translationType", s.config.Player.TranslationType)

	sources, err := s.animeClient.GetEpisodeSources(
		ctx,
		animeInfo.AllAnimeID,
		animeInfo.AllAnimeEpisodeNumber,
		s.config.Player.TranslationType,
	)

	if err != nil {
		return nil, fmt.Errorf("error fetching sources: %w", err)
	}

	log.Info("Retrieved all episode sources",
		"total_count", len(sources),
		"title", animeInfo.AllAnimeName,
		"episode", animeInfo.AllAnimeEpisodeNumber)

	// Filter sources to only include supported types (S-mp4 and Luf-mp4)
	var filteredSources []EpisodeSource
	for _, source := range sources {
		if strings.Contains(source.SourceName, "S-mp4") || strings.Contains(source.SourceName, "Luf-mp4") {
			filteredSources = append(filteredSources, source)
		}
	}

	log.Info("Filtered to supported sources",
		"supported_count", len(filteredSources),
		"filtered_out", len(sources)-len(filteredSources))

	if len(filteredSources) == 0 {
		log.Warn("No supported sources found for episode",
			"allAnimeID", animeInfo.AllAnimeID,
			"episodeNumber", animeInfo.AllAnimeEpisodeNumber)
		return nil, fmt.Errorf("no supported sources found for episode %s", animeInfo.AllAnimeEpisodeNumber)
	}

	// Sort sources by priority (highest first)
	sort.Slice(filteredSources, func(i, j int) bool {
		return filteredSources[i].Priority > filteredSources[j].Priority
	})

	return &EpisodeSourceInfo{
		AnimeName:       animeInfo.AllAnimeName,
		EpisodeNumber:   animeInfo.AllAnimeEpisodeNumber,
		AllAnimeID:      animeInfo.AllAnimeID,
		Sources:         filteredSources,
		TranslationType: s.config.Player.TranslationType,
	}, nil
}

// GetStreamURL decodes the source URL and fetches the actual streaming URL
func (s *PlayerService) GetStreamURL(ctx context.Context, source EpisodeSource) (string, error) {
	log.Debug("Getting stream URL for source", "sourceName", source.SourceName)

	// Decode the source URL
	decodedPath, err := s.decodeSourceURL(source.SourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to decode source URL: %w", err)
	}

	// Build the full API URL
	apiURL := "https://allanime.day" + decodedPath
	log.Debug("Decoded API URL", "url", apiURL)

	// Fetch the stream URL from the API
	streamURL, err := s.fetchStreamURL(ctx, apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch stream URL: %w", err)
	}

	log.Info("Retrieved stream URL", "sourceName", source.SourceName, "url", streamURL)
	return streamURL, nil
}

// decodeSourceURL decodes an encoded source URL from allanime
func (s *PlayerService) decodeSourceURL(encoded string) (string, error) {
	// Check if the string starts with "--"
	if len(encoded) < 2 || encoded[:2] != "--" {
		return "", fmt.Errorf("encoded string does not start with '--': %s", encoded)
	}

	// Remove the "--" prefix
	hexStr := encoded[2:]

	var decodedBuilder strings.Builder

	// Process each 2-character hex pair
	for i := 0; i < len(hexStr); i += 2 {
		if i+2 > len(hexStr) {
			return "", fmt.Errorf("invalid hex pair at position %d", i)
		}

		pair := hexStr[i : i+2]
		char := hexToChar(pair)

		if char == 0 {
			return "", fmt.Errorf("invalid hex pair: %s", pair)
		}

		decodedBuilder.WriteString(string(char))
	}

	decoded := decodedBuilder.String()

	// Replace "/clock" with "/clock.json" if needed
	decoded = strings.Replace(decoded, "/clock", "/clock.json", -1)

	return decoded, nil
}

// hexToChar maps hex pairs to their character representation
func hexToChar(pair string) rune {
	switch pair {
	case "01":
		return '9'
	case "08":
		return '0'
	case "05":
		return '='
	case "0a":
		return '2'
	case "0b":
		return '3'
	case "0c":
		return '4'
	case "07":
		return '?'
	case "00":
		return '8'
	case "5c":
		return 'd'
	case "0f":
		return '7'
	case "5e":
		return 'f'
	case "17":
		return '/'
	case "54":
		return 'l'
	case "09":
		return '1'
	case "48":
		return 'p'
	case "4f":
		return 'w'
	case "0e":
		return '6'
	case "5b":
		return 'c'
	case "5d":
		return 'e'
	case "0d":
		return '5'
	case "53":
		return 'k'
	case "1e":
		return '&'
	case "5a":
		return 'b'
	case "59":
		return 'a'
	case "4a":
		return 'r'
	case "4c":
		return 't'
	case "4e":
		return 'v'
	case "57":
		return 'o'
	case "51":
		return 'i'
	default:
		return 0
	}
}

// fetchStreamURL fetches the actual streaming URL from the decoded allanime URL
func (s *PlayerService) fetchStreamURL(ctx context.Context, url string) (string, error) {
	// Create an HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to mimic a browser
	req.Header.Set("User-Agent", allAnimeUserAgent)

	// Execute the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var response struct {
		Links []struct {
			Link string `json:"link"`
			HLS  bool   `json:"hls"`
		} `json:"links"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check if we have any links
	if len(response.Links) == 0 {
		return "", fmt.Errorf("no streaming links found in response")
	}

	// Return the first link (typically the best quality)
	return response.Links[0].Link, nil
}

// LaunchPlayer starts playback with the given stream URL and returns a channel for playback events
func (s *PlayerService) LaunchPlayer(ctx context.Context, streamURL string, episode AllAnimeEpisodeInfo) (<-chan PlaybackEvent, error) {
	log.Info("Launching media player",
		"player_type", s.config.Player.Type,
		"player_path", s.config.Player.Path)

	// Create the appropriate video player based on config
	videoPlayer, err := CreateVideoPlayer(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create video player: %w", err)
	}

	title := fmt.Sprintf("Ep %d - %s", episode.OverallEpisodeNumber, episode.PreferredTitle)

	// Start playback and get the events channel
	events, err := videoPlayer.Play(ctx, streamURL, title)
	if err != nil {
		return nil, fmt.Errorf("failed to start player: %w", err)
	}

	return events, nil
}

// parseArgs splits a string of command-line arguments, respecting quotes
func parseArgs(argsString string) []string {
	var args []string
	inQuotes := false
	current := ""

	for _, r := range argsString {
		switch r {
		case '"', '\'':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if current != "" {
					args = append(args, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}
