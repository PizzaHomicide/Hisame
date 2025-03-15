package service

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
)

type AnimeService struct {
	repo domain.AnimeRepository
	// TODO consider a map for faster access when looking for a specific anime by ID
	animeList []*domain.Anime // Keeps a local copy of all the anime, only updating it on user request
}

func NewAnimeService(repo domain.AnimeRepository) *AnimeService {
	return &AnimeService{
		repo: repo,
	}
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

// UpdateAnimeProgress updates an anime's progress and syncs with AniList
func (s *AnimeService) UpdateAnimeProgress(ctx context.Context, id int, progress int) error {
	// Find the anime in the local cache
	var anime *domain.Anime
	for _, a := range s.animeList {
		if a.ID == id {
			anime = a
			break
		}
	}

	if anime == nil || anime.UserData == nil {
		return fmt.Errorf("anime not found in list: %d", id)
	}

	anime.UserData.Progress = progress

	// Check if we should auto-complete
	if anime.Episodes > 0 && progress >= anime.Episodes {
		anime.UserData.Status = domain.StatusCompleted
	}

	return s.repo.UpdateUserAnimeData(ctx, id, anime.UserData)
}
