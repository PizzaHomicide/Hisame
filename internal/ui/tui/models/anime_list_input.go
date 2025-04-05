package models

// anime_list_input.go manages user input handling for the anime list view.
// It contains the main Update method to process tea.Msg events and delegates
// to appropriate handlers. This includes keyboard navigation, search input,
// and initiating playback actions based on user commands.

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"time"

	"github.com/PizzaHomicide/hisame/internal/log"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m *AnimeListModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in search mode, handle input differently
		if m.searchMode {
			switch msg.String() {
			case "esc":
				// Clear query
				log.Debug("Clearing search query")
				m.filters.searchQuery = ""
				m.searchInput.SetValue("")
				m.searchMode = false
				m.applyFilters()
				return m, nil
			case "enter":
				// Apply search
				log.Debug("Setting search query", "query", m.searchInput.Value())
				m.filters.searchQuery = m.searchInput.Value()
				m.searchMode = false
				m.applyFilters()
				return m, nil
			}

			// Let the text input handle other keys
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		// Normal mode key handling
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		if m.loading {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			return m, spinnerCmd
		}
		return m, nil

	case AnimeListMsg:
		if msg.Success {
			log.Debug("Anime list loaded")
			m.loading = false
			m.allAnime = m.animeService.GetAnimeList()
			m.applyFilters()
		} else {
			log.Debug("Anime list load error", "error", msg.Error)
			m.loading = false
			m.loadError = msg.Error
		}

	case AnimeUpdatedMsg:
		if msg.Success {
			log.Info("Anime updated successfully",
				"animeID", msg.AnimeID,
				"message", msg.Message)
			// Refresh the UI to show updated data
			m.applyFilters()
		} else {
			log.Error("Anime update failed",
				"animeID", msg.AnimeID,
				"error", msg.Error)
		}
		return m, nil

	case PlaybackCompletedMsg:
		if msg.Progress < 75.0 {
			log.Info("Playback ended.  Not incrementing progress as not enough of the episode was watched", "animeID", msg.AnimeID, "playbackProgress", msg.Progress)
			return m, nil
		}

		return m, func() tea.Msg {
			log.Info("Playback ended.  Incrementing progress", "animeID", msg.AnimeID, "playbackProgress", msg.Progress, "episode_watched", msg.EpisodeNumber)
			// Increment anime progress
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := m.animeService.IncrementProgress(ctx, msg.AnimeID)

			if err != nil {
				return AnimeUpdatedMsg{
					Success: false,
					AnimeID: msg.AnimeID,
					Error:   err,
				}
			}

			return AnimeUpdatedMsg{
				Success: true,
				AnimeID: msg.AnimeID,
				Message: fmt.Sprintf("Automatically updated progress after watching episode %d",
					msg.EpisodeNumber),
			}
		}

	}

	// Handle other message types in the playback file
	return m.handlePlaybackMessages(msg)
}

// handleKeyPress processes keyboard inputs in normal mode
func (m *AnimeListModel) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if len(m.filteredAnime) > 0 && m.cursor < len(m.filteredAnime)-1 {
			m.cursor++
		}
	case "1", "2", "3", "4", "5", "6":
		// Toggle status filters based on number keys
		m.toggleStatusFilter(msg.String())
		m.applyFilters()
		m.cursor = 0
	case "a":
		m.toggleHasNewEpisodesFilter()
		m.applyFilters()
		m.cursor = 0
	case "f":
		m.toggleIsFinishedAiringFilter()
		m.applyFilters()
		m.cursor = 0
	case "ctrl+f":
		m.searchMode = true
		m.searchInput.Focus()
		return m, nil
	case "enter":
		// TODO: Implement view detail of selected anime
		log.Info("View anime detail", "title", m.getSelectedAnime().Title.Preferred(m.config.UI.TitleLanguage), "id", m.getSelectedAnime().ID)
	case "p":
		return m.handlePlayEpisode()
	case "ctrl+p":
		return m.handleChooseEpisode()
	case "r":
		// Refresh anime list
		return m, func() tea.Msg {
			return LoadingMsg{
				Type:      LoadingStart,
				Message:   "Refreshing anime list...",
				Operation: m.fetchAnimeListCmd(),
			}
		}
	case "+":
		return m.handleIncrementProgress()
	case "-":
		return m.handleDecrementProgress()
	}

	return m, nil
}

// handleIncrementProgress handles incrementing the progress of the selected anime
func (m *AnimeListModel) handleIncrementProgress() (Model, tea.Cmd) {
	anime := m.getSelectedAnime()
	if anime == nil {
		return m, nil
	}

	return m, func() tea.Msg {
		log.Info("Incrementing progress",
			"title", anime.Title.Preferred(m.config.UI.TitleLanguage),
			"id", anime.ID,
			"current_progress", anime.UserData.Progress)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.animeService.IncrementProgress(ctx, anime.ID)
		if err != nil {
			log.Error("Failed to increment progress", "error", err)
			return AnimeUpdatedMsg{
				Success: false,
				AnimeID: anime.ID,
				Error:   err,
			}
		}

		return AnimeUpdatedMsg{
			Success: true,
			AnimeID: anime.ID,
			Message: fmt.Sprintf("Updated progress for %s to %d/%d",
				anime.Title.Preferred(m.config.UI.TitleLanguage),
				anime.UserData.Progress,
				anime.Episodes),
		}
	}
}

// handleDecrementProgress handles decrementing the progress of the selected anime
func (m *AnimeListModel) handleDecrementProgress() (Model, tea.Cmd) {
	anime := m.getSelectedAnime()
	if anime == nil {
		return m, nil
	}

	return m, func() tea.Msg {
		log.Info("Decrementing progress",
			"title", anime.Title.Preferred(m.config.UI.TitleLanguage),
			"id", anime.ID,
			"current_progress", anime.UserData.Progress)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.animeService.DecrementProgress(ctx, anime.ID)
		if err != nil {
			log.Error("Failed to decrement progress", "error", err)
			return AnimeUpdatedMsg{
				Success: false,
				AnimeID: anime.ID,
				Error:   err,
			}
		}

		return AnimeUpdatedMsg{
			Success: true,
			AnimeID: anime.ID,
			Message: fmt.Sprintf("Updated progress for %s to %d/%d",
				anime.Title.Preferred(m.config.UI.TitleLanguage),
				anime.UserData.Progress,
				anime.Episodes),
		}
	}
}

// handlePlayEpisode initiates playback of the next episode
func (m *AnimeListModel) handlePlayEpisode() (Model, tea.Cmd) {
	nextEpNumber := m.getSelectedAnime().UserData.Progress + 1
	log.Info("Play next episode",
		"title", m.getSelectedAnime().Title.Preferred(m.config.UI.TitleLanguage),
		"id", m.getSelectedAnime().ID,
		"current_progress", m.getSelectedAnime().UserData.Progress,
		"next_ep", nextEpNumber)

	// Set loading state with custom message
	m.loading = true
	m.loadingMsg = fmt.Sprintf("Finding episode %d for %s...",
		nextEpNumber,
		m.getSelectedAnime().Title.Preferred(m.config.UI.TitleLanguage))

	return m, tea.Batch(
		m.spinner.Tick,
		m.loadNextEpisode(nextEpNumber),
	)
}

// handleChooseEpisode initiates the episode selection flow
func (m *AnimeListModel) handleChooseEpisode() (Model, tea.Cmd) {
	log.Info("Choose episode to play",
		"title", m.getSelectedAnime().Title.Preferred(m.config.UI.TitleLanguage),
		"id", m.getSelectedAnime().ID)

	m.loading = true
	m.loadingMsg = fmt.Sprintf("Finding episodes for %s...",
		m.getSelectedAnime().Title.Preferred(m.config.UI.TitleLanguage))

	return m, tea.Batch(
		m.spinner.Tick,
		m.loadEpisodes(),
	)
}
