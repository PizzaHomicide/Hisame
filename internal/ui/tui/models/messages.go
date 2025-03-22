package models

import "github.com/PizzaHomicide/hisame/internal/player"

// AuthCompletedMsg is sent when authentication is completed successfully
type AuthCompletedMsg struct {
	Token string
}

// AuthFailedMsg is sent when authentication fails
type AuthFailedMsg struct {
	Error string
}

// AnimeListLoadedMsg is sent when the anime list is loaded
type AnimeListLoadedMsg struct{}

// AnimeListErrorMsg is sent when there's an error loading the anime list
type AnimeListErrorMsg struct {
	Error error
}

type EpisodeLoadedMsg struct {
	Episodes []player.AllAnimeEpisodeInfo
	Title    string
}

type EpisodeLoadErrorMsg struct {
	Error error
}

type EpisodeSelectMsg struct {
	Episode *player.AllAnimeEpisodeInfo
}

type NextEpisodeFoundMsg struct {
	Episode player.AllAnimeEpisodeInfo
}

type EpisodeSourcesLoadedMsg struct {
	Sources     *player.EpisodeSourceInfo
	EpisodeInfo player.AllAnimeEpisodeInfo
}

type EpisodeSourcesErrorMsg struct {
	Error       error
	EpisodeInfo player.AllAnimeEpisodeInfo
}
