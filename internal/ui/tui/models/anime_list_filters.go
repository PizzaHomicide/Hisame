package models

// anime_list_filters.go handles all anime list filtering functionality.
// It contains methods for toggling different filter types (status, airing, etc.),
// applying filters to the anime list, and rendering the current filter status.

import (
	"fmt"
	"strings"

	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"

	"slices"

	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
)

// toggleFilter toggles a filter based on the action
func (m *AnimeListModel) toggleFilter(action kb.Action) {
	var status domain.MediaStatus

	switch action {
	case kb.ActionToggleFilterStatusCurrent:
		status = domain.StatusCurrent
	case kb.ActionToggleFilterStatusPlanning:
		status = domain.StatusPlanning
	case kb.ActionToggleFilterStatusComplete:
		status = domain.StatusCompleted
	case kb.ActionToggleFilterStatusDropped:
		status = domain.StatusDropped
	case kb.ActionToggleFilterStatusPaused:
		status = domain.StatusPaused
	case kb.ActionToggleFilterStatusRepeating:
		status = domain.StatusRepeating
	case kb.ActionToggleFilterFinishedAiring:
		m.filters.isFinishedAiring = !m.filters.isFinishedAiring
		return
	case kb.ActionToggleFilterNewEpisodes:
		m.filters.hasAvailableEpisodes = !m.filters.hasAvailableEpisodes
		return
	default:
		return
	}

	// Check if the status is already in the filters
	index := -1
	for i, s := range m.filters.statusFilters {
		if s == status {
			index = i
			break
		}
	}

	if index >= 0 {
		// Status is already in filters, remove it
		m.filters.statusFilters = slices.Delete(m.filters.statusFilters, index, index+1)

		// If we removed all filters, use the defaults
		if len(m.filters.statusFilters) == 0 {
			m.filters.statusFilters = DEFAULT_STATUS_FILTERS
		}
	} else {
		// Status not in filters, add it
		m.filters.statusFilters = append(m.filters.statusFilters, status)
	}
}

// applyFilters applies the current filters to the anime list
func (m *AnimeListModel) applyFilters() {
	// Start with all anime that match status filters
	statusFilteredAnime := []*domain.Anime{}

	// Apply status filters
	for _, anime := range m.allAnime {
		if anime.UserData == nil {
			continue
		}

		// Check if the anime's status is in our status filters
		statusMatch := slices.Contains(m.filters.statusFilters, anime.UserData.Status)

		if statusMatch {
			statusFilteredAnime = append(statusFilteredAnime, anime)
		}
	}

	// Apply additional filters if needed
	m.filteredAnime = []*domain.Anime{}

	for _, anime := range statusFilteredAnime {
		includeAnime := true

		// Filter for has new episodes if enabled
		if m.filters.hasAvailableEpisodes {
			if !anime.HasUnwatchedEpisodes() {
				includeAnime = false
			}
		}

		// Filter for completed airing if enabled
		if m.filters.isFinishedAiring && includeAnime {
			// Check if the anime has finished airing
			log.Debug("Anime status..", "title", anime.Title.English, "status", anime.Status)
			isComplete := anime.Status == "FINISHED"
			if !isComplete {
				includeAnime = false
			}
		}

		// Filter on title search query
		if m.filters.searchQuery != "" && includeAnime {
			query := strings.ToLower(m.filters.searchQuery)

			// Check only the current anime being processed
			title := strings.ToLower(anime.Title.Preferred)
			if !strings.Contains(title, query) {
				includeAnime = false
			}
		}

		if includeAnime {
			m.filteredAnime = append(m.filteredAnime, anime)
		}
	}

	// Reset cursor if it's out of bounds
	if len(m.filteredAnime) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filteredAnime) {
		m.cursor = len(m.filteredAnime) - 1
	}
}

// getStatusFilterCounts returns a map with the count of anime for each status
func (m *AnimeListModel) getStatusFilterCounts() map[domain.MediaStatus]int {
	counts := make(map[domain.MediaStatus]int)

	statuses := []domain.MediaStatus{
		domain.StatusCurrent,
		domain.StatusPlanning,
		domain.StatusCompleted,
		domain.StatusDropped,
		domain.StatusPaused,
	}

	// Initialize all counts to 0
	for _, status := range statuses {
		counts[status] = 0
	}

	// Count anime by status
	for _, anime := range m.allAnime {
		if anime.UserData != nil {
			counts[anime.UserData.Status]++
		}
	}

	return counts
}

// renderFilterStatus returns a concise string representation of all active filters
func (m *AnimeListModel) renderFilterStatus() string {
	// Status filters
	statusFilters := []struct {
		status    domain.MediaStatus
		indicator string
	}{
		{domain.StatusCurrent, "W"},
		{domain.StatusPlanning, "P"},
		{domain.StatusCompleted, "C"},
		{domain.StatusDropped, "D"},
		{domain.StatusPaused, "H"},
		{domain.StatusRepeating, "R"},
	}

	// Create status filter indicators
	var statusIndicators []string
	for _, s := range statusFilters {
		// Check if this status is in the active filters
		isActive := false
		for _, activeStatus := range m.filters.statusFilters {
			if activeStatus == s.status {
				isActive = true
				break
			}
		}

		// Format the indicator based on active status
		if isActive {
			statusIndicators = append(statusIndicators, fmt.Sprintf("[%s]", s.indicator))
		} else {
			statusIndicators = append(statusIndicators, "[-]")
		}
	}

	episodeFilters := fmt.Sprintf("| Episodes -> [%s] [%s]",
		conditionalIndicator(m.filters.hasAvailableEpisodes, "A", "-"),
		conditionalIndicator(m.filters.isFinishedAiring, "F", "-"))

	searchText := "-"
	if m.filters.searchQuery != "" {
		searchText = fmt.Sprintf("\"%s\"", m.filters.searchQuery)
	}
	searchFilter := fmt.Sprintf(" | Search: %s", searchText)

	// Join all filter sections
	filterLine := " Status -> " + strings.Join(statusIndicators, " ") + " " + episodeFilters + " " + searchFilter
	filterPrefix := styles.Title.Render("Filters:")
	return filterPrefix + styles.FilterStatus.Render(filterLine)
}

// Helper function to return the appropriate indicator based on a condition
func conditionalIndicator(condition bool, activeChar, inactiveChar string) string {
	if condition {
		return activeChar
	}
	return inactiveChar
}
