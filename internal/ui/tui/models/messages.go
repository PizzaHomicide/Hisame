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
	StreamURL   string
}

type EpisodeSourcesErrorMsg struct {
	Error       error
	EpisodeInfo player.AllAnimeEpisodeInfo
}

// PlaybackStartedMsg is sent when MPV playback has successfully started
type PlaybackStartedMsg struct {
	EpisodeInfo player.AllAnimeEpisodeInfo
}

// PlaybackEndedMsg is sent when MPV playback has ended
type PlaybackEndedMsg struct {
	EpisodeInfo player.AllAnimeEpisodeInfo
	Progress    float64 // Percentage of playback completed
}

// PlaybackProgressMsg is sent to provide updates on playback progress
type PlaybackProgressMsg struct {
	EpisodeInfo player.AllAnimeEpisodeInfo
	Progress    float64 // Percentage of playback completed
}

type PlaybackErrorMsg struct {
	Error       error
	EpisodeInfo player.AllAnimeEpisodeInfo
}
