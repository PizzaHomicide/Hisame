package models

// anime_list_input.go manages user input handling for the anime list view.
// It contains the main Update method to process tea.Msg events and delegates
// to appropriate handlers. This includes keyboard navigation, search input,
// and initiating playback actions based on user commands.

import (
	"context"
	"fmt"
	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
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
		if cmd := m.handleSearchModeKeyMsg(msg); cmd != nil {
			return m, cmd
		}

		// Normal mode key handling
		if cmd := m.handleKeyPress(msg); cmd != nil {
			return m, cmd
		}

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

func (m *AnimeListModel) handleSearchModeKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if !m.searchMode {
		return nil
	}
	switch kb.GetActionByKey(msg, kb.ContextSearchMode) {
	case kb.ActionBack:
		// Cancels search, clearing the filter
		m.searchMode = false
		m.filters.searchQuery = "" // TODO: This seems redundant, align with how episode_select filter works
		m.searchInput.SetValue("")
		m.applyFilters()
		return Handled("search:exit")
	case kb.ActionSearchComplete:
		m.searchMode = false
		m.filters.searchQuery = m.searchInput.Value()
		m.applyFilters()
		return Handled("search:apply")
	}

	// Let the text input model handle other keys
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Apply filters as we type
	m.filters.searchQuery = m.searchInput.Value()
	m.applyFilters()

	return cmd
}

// handleKeyPress processes keyboard inputs in normal mode
func (m *AnimeListModel) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch action := kb.GetActionByKey(msg, kb.ContextAnimeList); action {
	case kb.ActionMoveUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return Handled("cursor_move:up")
	case kb.ActionMoveDown:
		if len(m.filteredAnime) > 0 && m.cursor < len(m.filteredAnime)-1 {
			m.cursor++
		}
		return Handled("cursor_move:down")
	// All filter toggle actions are handled together
	case kb.ActionToggleFilterStatusCurrent, kb.ActionToggleFilterStatusPlanning, kb.ActionToggleFilterStatusComplete,
		kb.ActionToggleFilterStatusDropped, kb.ActionToggleFilterStatusPaused, kb.ActionToggleFilterStatusRepeating,
		kb.ActionToggleFilterFinishedAiring, kb.ActionToggleFilterNewEpisodes:
		m.toggleFilter(action)
		m.applyFilters()
		m.cursor = 0
		return Handled("filter:toggle")
	case kb.ActionEnableSearch:
		m.searchMode = true
		m.searchInput.Focus()
		return Handled("search:enable")
	case kb.ActionPlayNextEpisode:
		return m.handlePlayEpisode()
	case kb.ActionOpenEpisodeSelector:
		return m.handleChooseEpisode()
	case kb.ActionRefreshAnimeList:
		return func() tea.Msg {
			return LoadingMsg{
				Type:      LoadingStart,
				Message:   "Refreshing anime list...",
				Operation: m.fetchAnimeListCmd(),
			}
		}
	case kb.ActionIncrementProgress:
		return m.handleIncrementProgress()
	case kb.ActionDecrementProgress:
		return m.handleDecrementProgress()
	case kb.ActionViewAnimeDetails:
		anime := m.getSelectedAnime()
		if anime == nil {
			return Handled("view_anime_details:none_selected")
		}
		return func() tea.Msg {
			return AnimeDetailsMsg{
				Anime: anime,
			}
		}
	}

	return nil
}

// handleIncrementProgress handles incrementing the progress of the selected anime
func (m *AnimeListModel) handleIncrementProgress() tea.Cmd {
	anime := m.getSelectedAnime()
	if anime == nil {
		return Handled("increment_progress:none_selected")
	}

	return func() tea.Msg {
		log.Info("Incrementing progress",
			"title", anime.Title.Preferred,
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
				anime.Title.Preferred,
				anime.UserData.Progress,
				anime.Episodes),
		}
	}
}

// handleDecrementProgress handles decrementing the progress of the selected anime
func (m *AnimeListModel) handleDecrementProgress() tea.Cmd {
	anime := m.getSelectedAnime()
	if anime == nil {
		return Handled("decrement_progress:none_selected")
	}

	return func() tea.Msg {
		log.Info("Decrementing progress",
			"title", anime.Title.Preferred,
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
				anime.Title.Preferred,
				anime.UserData.Progress,
				anime.Episodes),
		}
	}
}

// handlePlayEpisode initiates playback of the next episode
func (m *AnimeListModel) handlePlayEpisode() tea.Cmd {
	// Only attempt playback if there are unwatched episodes available
	anime := m.getSelectedAnime()
	if !anime.HasUnwatchedEpisodes() {
		log.Info("No unwatched episodes available", "title", anime.Title.Preferred,
			"id", anime.ID, "progress", anime.UserData.Progress, "latest_aired", anime.GetLatestAiredEpisode())
		return Handled("play_episode:none_available")
	}
	nextEpNumber := m.getSelectedAnime().UserData.Progress + 1
	log.Info("Play next episode",
		"title", m.getSelectedAnime().Title.Preferred,
		"id", m.getSelectedAnime().ID,
		"current_progress", m.getSelectedAnime().UserData.Progress,
		"next_ep", nextEpNumber)

	// Set loading state with custom message
	m.loading = true
	m.loadingMsg = fmt.Sprintf("Finding episode %d for %s...",
		nextEpNumber,
		m.getSelectedAnime().Title.Preferred)

	return tea.Batch(
		m.spinner.Tick,
		m.loadNextEpisode(nextEpNumber),
	)
}

// handleChooseEpisode initiates the episode selection flow
func (m *AnimeListModel) handleChooseEpisode() tea.Cmd {
	log.Info("Choose episode to play",
		"title", m.getSelectedAnime().Title.Preferred,
		"id", m.getSelectedAnime().ID)

	m.loading = true
	m.loadingMsg = fmt.Sprintf("Finding episodes for %s...",
		m.getSelectedAnime().Title.Preferred)

	return tea.Batch(
		m.spinner.Tick,
		m.loadEpisodes(),
	)
}
