package player

import (
	"context"
	"errors"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	"sort"
	"strconv"
	"strings"
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
func (s *PlayerService) FindEpisodes(ctx context.Context, animeID int, title string, synonyms []string) (*FindEpisodesResult, error) {
	log.Debug("Finding episodes", "title", title, "id", animeID, "synonyms", synonyms)

	// Search for shows matching the anime title
	shows, err := s.animeClient.SearchShows(ctx, title, s.config.Player.TranslationType)
	if err != nil {
		return nil, fmt.Errorf("error searching for shows: %w", err)
	}

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
	result := s.buildEpisodeList(matchedShows, animeID)

	log.Debug("Built episode list", "matched_show_count", len(matchedShows), "episode_count", len(result.Episodes), "title", title)

	return result, nil
}

// matchesByTitleOrSynonyms checks if a show matches the anime by title or synonyms
func (s *PlayerService) matchesByTitleOrSynonyms(title string, synonyms []string, show AllAnimeShow) bool {
	// Check if the anime title matches any of the show's names
	animeTitle := strings.ToLower(title)
	if strings.ToLower(show.Name) == animeTitle ||
		strings.ToLower(show.EnglishName) == animeTitle ||
		strings.ToLower(show.NativeName) == animeTitle {
		log.Debug("Title match found", "title", title, "allanime_name", show.Name,
			"allanime_englishname", show.EnglishName, "allanime_nativename", show.NativeName)
		return true
	}

	// Check if any of the show's alt names match any of the anime's synonyms
	for _, altName := range show.TrustedAltNames {
		altNameLower := strings.ToLower(altName)

		// Check against anime title
		if altNameLower == animeTitle {
			log.Debug("Alt name match found", "alt_name", altName, "title", title)
			return true
		}

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
func (s *PlayerService) buildEpisodeList(shows []AllAnimeShow, animeID int) *FindEpisodesResult {
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
				Title:                 show.Name,
				EnglishTitle:          show.EnglishName,
				NativeTitle:           show.NativeName,
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
