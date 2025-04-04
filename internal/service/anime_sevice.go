package service

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
	"sync"
)

type AnimeService struct {
	repo domain.AnimeRepository
	// TODO consider a map for faster access when looking for a specific anime by ID
	animeList  []*domain.Anime // Keeps a local copy of all the anime, only updating it on user request
	updateLock sync.Mutex
}

func NewAnimeService(repo domain.AnimeRepository) *AnimeService {
	return &AnimeService{
		repo: repo,
	}
}

func (s *AnimeService) GetAnimeList() []*domain.Anime {
	return s.animeList
}

// LoadAnimeList fetches the complete anime list from the repository
func (s *AnimeService) LoadAnimeList(ctx context.Context) error {
	list, err := s.repo.GetAllAnimeList(ctx)
	if err != nil {
		return err
	}

	s.animeList = list
	return nil
}

// GetAnimeListByStatus filters the cached anime list by status
func (s *AnimeService) GetAnimeListByStatus(status domain.MediaStatus) []*domain.Anime {
	var result []*domain.Anime

	for _, anime := range s.animeList {
		if anime.UserData != nil && anime.UserData.Status == status {
			result = append(result, anime)
		}
	}

	return result
}

// GetAnimeByID finds an anime in the cached list by its ID
func (s *AnimeService) GetAnimeByID(id int) *domain.Anime {
	for _, anime := range s.animeList {
		if anime.ID == id {
			return anime
		}
	}
	return nil
}

// IncrementProgress increases the progress for an anime by 1
// Returns an error if progress is already at or above episode count
func (s *AnimeService) IncrementProgress(ctx context.Context, animeID int) error {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	// Find the anime in our cached list
	anime := s.GetAnimeByID(animeID)
	if anime == nil {
		return fmt.Errorf("anime not found with ID: %d", animeID)
	}

	// Get current values
	currentProgress := anime.UserData.Progress
	totalEpisodes := anime.Episodes

	// Validate if we can increment
	if totalEpisodes > 0 && currentProgress >= totalEpisodes {
		return fmt.Errorf("cannot increment progress: already completed all %d episodes", totalEpisodes)
	}

	// Calculate new progress
	newProgress := currentProgress + 1

	// Create update parameters
	progressValue := newProgress // Using a variable because we need its address
	params := &domain.AnimeUpdateParams{
		MediaID:  animeID,
		Progress: &progressValue,
	}

	// Send update to repository
	result, err := s.repo.UpdateAnime(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	s.syncAnimeWithUpdateResult(anime, result)

	// Log basic info about the update
	log.Info("Incremented anime progress",
		"animeID", animeID,
		"title", anime.Title.Preferred("english"),
		"progress", fmt.Sprintf("%d/%d", result.Progress, totalEpisodes),
		"status", result.Status)

	// Log special messages for starting or completing
	if currentProgress == 0 && newProgress == 1 {
		log.Info("Started watching a new anime",
			"title", anime.Title.Preferred("english"),
			"id", animeID)
	}

	if totalEpisodes > 0 && newProgress == totalEpisodes {
		log.Info("Completed all episodes of anime",
			"title", anime.Title.Preferred("english"),
			"id", animeID,
			"episodes", totalEpisodes)
	}

	return nil
}

// DecrementProgress decreases the progress for an anime by 1
// Returns an error if progress is already 0
func (s *AnimeService) DecrementProgress(ctx context.Context, animeID int) error {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	// Find the anime in our cached list
	anime := s.GetAnimeByID(animeID)
	if anime == nil {
		return fmt.Errorf("anime not found with ID: %d", animeID)
	}

	// Get current values
	currentProgress := anime.UserData.Progress
	totalEpisodes := anime.Episodes

	// Validate if we can decrement
	if currentProgress <= 0 {
		return fmt.Errorf("cannot decrement progress: already at 0 episodes")
	}

	// Calculate new progress
	newProgress := currentProgress - 1

	// Create update parameters
	progressValue := newProgress // Using a variable because we need its address
	params := &domain.AnimeUpdateParams{
		MediaID:  animeID,
		Progress: &progressValue,
	}

	// Send update to repository
	result, err := s.repo.UpdateAnime(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	s.syncAnimeWithUpdateResult(anime, result)

	// Log basic info about the update
	log.Info("Decremented anime progress",
		"animeID", animeID,
		"title", anime.Title.Preferred("english"),
		"progress", fmt.Sprintf("%d/%d", result.Progress, totalEpisodes),
		"status", result.Status)

	// Special case - if anime was completed and is now un-completed
	previouslyCompleted := anime.UserData.Status == domain.StatusCompleted &&
		currentProgress == totalEpisodes &&
		newProgress < totalEpisodes

	if previouslyCompleted && result.Status != domain.StatusCompleted {
		log.Info("Anime un-completed",
			"title", anime.Title.Preferred("english"),
			"id", animeID,
			"new_status", result.Status)
	}

	return nil
}

// syncAnimeWithUpdateResult updates the cached anime data with values from an update result
func (s *AnimeService) syncAnimeWithUpdateResult(anime *domain.Anime, result *domain.AnimeUpdateResult) {
	if anime == nil || result == nil || anime.UserData == nil {
		return
	}

	// Update standard fields
	anime.UserData.Status = result.Status
	anime.UserData.Progress = result.Progress
	anime.UserData.Score = result.Score
	anime.UserData.Notes = result.Notes
	anime.UserData.StartDate = result.StartDate
	anime.UserData.EndDate = result.CompletionDate

	log.Debug("Synchronized local anime data with update result",
		"animeID", anime.ID,
		"title", anime.Title.Preferred("english"),
		"status", result.Status,
		"progress", result.Progress)
}
