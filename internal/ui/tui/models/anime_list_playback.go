package models

// anime_list_playback.go encapsulates all functionality related to episode playback.
// It contains methods for finding episodes, loading sources, launching the media player,
// and handling playback-related messages.

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"time"

	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePlaybackMessages handles all playback-related messages
func (m *AnimeListModel) handlePlaybackMessages(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PlaybackMsg:
		switch msg.Type {
		case PlaybackEventEpisodeFound:
			log.Info("Next episode found, loading sources",
				"title", msg.Anime.Title.Preferred(m.config.UI.TitleLanguage),
				"overall_epNum", msg.Episode.OverallEpisodeNumber,
				"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
				"allanime_id", msg.Episode.AllAnimeID,
				"anilist_id", msg.Anime.ID)

			// Start loading the sources for this episode
			m.loading = true
			m.loadingMsg = fmt.Sprintf("Loading sources for episode %s of %s...",
				msg.Episode.AllAnimeEpisodeNumber,
				msg.Episode.Title)

			return m, tea.Batch(
				m.spinner.Tick,
				m.playEpisode(msg.Episode, msg.Anime),
			)

		case PlaybackEventSourcesLoaded:
			m.loading = false

			log.Info("Episode sources loaded successfully",
				"title", msg.Episode.Title,
				"episode", msg.Episode.AllAnimeEpisodeNumber,
				"source_count", len(msg.Sources.Sources))

			// Log details about each source
			for i, source := range msg.Sources.Sources {
				log.Debug("Source option",
					"index", i,
					"name", source.SourceName,
					"priority", source.Priority,
					"type", source.Type,
					"has_download", source.Downloads != nil)
			}

			// At this point, we would normally launch the player
			// For now, just log that we would play the highest priority source
			if len(msg.Sources.Sources) > 0 {
				bestSource := msg.Sources.Sources[0] // Already sorted by priority
				log.Info("Would play this source (highest priority)",
					"name", bestSource.SourceName,
					"priority", bestSource.Priority,
					"type", bestSource.Type,
					"url", msg.StreamURL)
			}

			return m, nil

		case PlaybackEventError:
			m.loading = false

			log.Error("Failed to load episode sources",
				"title", msg.Episode.Title,
				"episode", msg.Episode.AllAnimeEpisodeNumber,
				"error", msg.Error)

			return m, nil

		case PlaybackEventStarted:
			m.loading = false
			log.Info("Playback started",
				"title", msg.Episode.Title,
				"episode", msg.Episode.AllAnimeEpisodeNumber)
			return m, m.listenForPlaybackCompletion()

		case PlaybackEventEnded:
			m.loading = false
			log.Info("Playback ended",
				"title", msg.Episode.Title,
				"episode", msg.Episode.AllAnimeEpisodeNumber,
				"progress", msg.Progress)
			return m, nil

		case PlaybackEventProgress:
			log.Debug("Playback progress",
				"title", msg.Episode.Title,
				"episode", msg.Episode.AllAnimeEpisodeNumber,
				"progress", msg.Progress)
			return m, nil
		}

	case EpisodeMsg:
		switch msg.Type {
		case EpisodeEventSelected:
			if msg.Episode != nil {
				log.Info("Episode selected from modal",
					"overall_epNum", msg.Episode.OverallEpisodeNumber,
					"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
					"allanime_id", msg.Episode.AllAnimeID,
					"title", msg.Episode.Title)

				// Start loading the sources
				m.loading = true
				m.loadingMsg = fmt.Sprintf("Loading sources for episode %s of %s...",
					msg.Episode.AllAnimeEpisodeNumber,
					msg.Episode.Title)

				return m, tea.Batch(
					m.spinner.Tick,
					m.playEpisode(*msg.Episode, nil),
				)
			}
		}
	}

	return m, nil
}

// loadEpisodes loads all episodes for the selected anime
func (m *AnimeListModel) loadEpisodes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anime := m.getSelectedAnime()

		epResult, err := m.playerService.FindEpisodes(
			ctx,
			anime.ID,
			anime.Title.Native,
			anime.Synonyms,
		)

		if err != nil {
			log.Error("Failed to get episodes", "error", err)
			return EpisodeMsg{
				Type:  EpisodeEventError,
				Error: err,
			}
		}

		return EpisodeMsg{
			Type:     EpisodeEventLoaded,
			Episodes: epResult.Episodes,
			Title:    anime.Title.Preferred(m.config.UI.TitleLanguage),
		}
	}
}

// loadNextEpisode loads the specific next episode for an anime
func (m *AnimeListModel) loadNextEpisode(nextEpNumber int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anime := m.getSelectedAnime()

		eps, err := m.playerService.FindEpisodes(
			ctx,
			anime.ID,
			anime.Title.Native,
			anime.Synonyms,
		)

		if err != nil {
			log.Error("Failed to get episodes", "error", err)
			return EpisodeMsg{
				Type:  EpisodeEventError,
				Error: err,
			}
		}

		// Find the specific episode we want
		var selectedEp *player.AllAnimeEpisodeInfo
		for i, ep := range eps.Episodes {
			if ep.OverallEpisodeNumber == nextEpNumber {
				selectedEp = &eps.Episodes[i]
				break
			}
		}

		if selectedEp == nil {
			log.Error("Could not find next episode", "nextEp", nextEpNumber)
			return EpisodeMsg{
				Type:  EpisodeEventError,
				Error: fmt.Errorf("could not find episode %d", nextEpNumber),
			}
		}

		// Success! Return the found episode
		log.Info("Selected next episode to play",
			"overall_epNum", selectedEp.OverallEpisodeNumber,
			"allanime_epNum", selectedEp.AllAnimeEpisodeNumber,
			"allanime_id", selectedEp.AllAnimeID,
			"anilist_id", selectedEp.AniListID)

		return PlaybackMsg{
			Type:    PlaybackEventEpisodeFound,
			Episode: *selectedEp,
			Anime:   anime,
		}
	}
}

// playEpisode attempts to play the given episode.  Use nil `anime` to skip automatic progress updates
func (m *AnimeListModel) playEpisode(episode player.AllAnimeEpisodeInfo, anime *domain.Anime) tea.Cmd {
	return func() tea.Msg {
		// Create a context with timeout for the entire operation
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel() // This ensures the main context is always canceled

		// Set loading state for source fetching
		log.Info("Fetching sources for episode",
			"title", episode.Title,
			"overall_epNum", episode.OverallEpisodeNumber,
			"allanime_epNum", episode.AllAnimeEpisodeNumber)

		// Get sources for the episode
		sources, err := m.playerService.GetEpisodeSources(ctx, episode)
		if err != nil {
			log.Error("Failed to get episode sources", "error", err)
			return PlaybackMsg{
				Type:    PlaybackEventError,
				Error:   err,
				Episode: episode,
			}
		}

		// Try to get a working stream URL from each source until one works
		var streamURL string
		var successSource player.EpisodeSource

		for _, source := range sources.Sources {
			log.Info("Attempting to get stream URL",
				"source_name", source.SourceName,
				"priority", source.Priority)

			url, err := m.playerService.GetStreamURL(ctx, source)
			if err != nil {
				log.Warn("Failed to get stream URL from source",
					"source_name", source.SourceName,
					"error", err)
				continue // Try the next source
			}

			// Success!
			streamURL = url
			successSource = source
			break
		}

		if streamURL == "" {
			return PlaybackMsg{
				Type:    PlaybackEventError,
				Error:   fmt.Errorf("failed to get playable URL from any source"),
				Episode: episode,
			}
		}

		// Log the URL that would be used to play the episode
		log.Info("Found playable stream URL",
			"source_name", successSource.SourceName)

		// Update loading message to indicate we're starting the player
		m.loadingMsg = fmt.Sprintf("Launching media player for %s episode %s...",
			episode.Title, episode.AllAnimeEpisodeNumber)

		// Create a new context for the playback monitoring that's independent of this function
		playbackCtx, playbackCancel := context.WithCancel(context.Background())

		// Launch the player with the stream URL and get the event channel
		eventCh, err := m.playerService.LaunchPlayer(playbackCtx, streamURL)
		if err != nil {
			playbackCancel() // Clean up the playback context if launch fails
			log.Error("Failed to launch media player", "error", err)
			return PlaybackMsg{
				Type:    PlaybackEventError,
				Error:   fmt.Errorf("failed to launch player: %w", err),
				Episode: episode,
			}
		}

		// Update loading message to indicate we're waiting for playback to start
		m.loadingMsg = fmt.Sprintf("Waiting for playback to start for %s episode %s...",
			episode.Title, episode.AllAnimeEpisodeNumber)

		// Wait for the first event (should be playback started or an error)
		select {
		case <-ctx.Done():
			playbackCancel() // Clean up the playback context on timeout
			return PlaybackMsg{
				Type:    PlaybackEventError,
				Error:   fmt.Errorf("timeout waiting for playback to start"),
				Episode: episode,
			}
		case event, ok := <-eventCh:
			if !ok {
				playbackCancel() // Clean up the playback context on channel close
				return PlaybackMsg{
					Type:    PlaybackEventError,
					Error:   fmt.Errorf("player event channel closed unexpectedly"),
					Episode: episode,
				}
			}

			// Handle the event based on its type
			switch event.Type {
			case player.PlaybackStarted:
				log.Info("MPV playback started successfully")

				// Start another goroutine to continue monitoring playback progress
				go func() {
					defer playbackCancel() // Ensure context is canceled when goroutine exits

					for event := range eventCh {
						switch event.Type {
						case player.PlaybackEnded:
							log.Info("MPV playback ended", "progress", event.Progress)
							m.playbackCompletionCh <- PlaybackCompletedMsg{
								AnimeID:       anime.ID,
								EpisodeNumber: episode.OverallEpisodeNumber,
								Progress:      event.Progress,
							}
							return
						case player.PlaybackError:
							log.Error("MPV playback error", "error", event.Error)
							return
						}
					}
					log.Debug("MPV event channel closed, stopping monitoring")
				}()

				// Return a message indicating playback has started
				return PlaybackMsg{
					Type:    PlaybackEventStarted,
					Episode: episode,
				}

			case player.PlaybackError:
				playbackCancel() // Clean up the playback context on error
				log.Error("MPV failed to start playback", "error", event.Error)
				return PlaybackMsg{
					Type:    PlaybackEventError,
					Error:   event.Error,
					Episode: episode,
				}
			default:
				// TODO:  I don't think I want this.  Let's just report an error playback message, but indicate it _may_ have worked, but monitoring will be unavailable.
				log.Warn("Unexpected initial event from MPV", "event_type", event.Type)
				// Treat as started anyway to be safe
				go func() {
					defer playbackCancel() // Ensure context is canceled when goroutine exits

					for event := range eventCh {
						switch event.Type {
						case player.PlaybackEnded:
							log.Info("MPV playback ended")
							return
						case player.PlaybackError:
							log.Error("MPV playback error", "error", event.Error)
							return
						}
					}
				}()
				return PlaybackMsg{
					Type:    PlaybackEventStarted,
					Episode: episode,
				}
			}
		}
	}
}

func (m *AnimeListModel) listenForPlaybackCompletion() tea.Cmd {
	return func() tea.Msg {
		event := <-m.playbackCompletionCh
		return event
	}
}
