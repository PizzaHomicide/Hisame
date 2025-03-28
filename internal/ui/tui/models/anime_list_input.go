package models

// anime_list_input.go manages user input handling for the anime list view.
// It contains the main Update method to process tea.Msg events and delegates
// to appropriate handlers. This includes keyboard navigation, search input,
// and initiating playback actions based on user commands.

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"

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
		m.loading = true
		m.loadingMsg = "Loading anime list..."
		m.loadError = nil
		return m, tea.Batch(
			m.spinner.Tick,
			loadAnimeList(m.animeService),
		)
	}

	return m, nil
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
