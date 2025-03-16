package models

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/mattn/go-runewidth"
	"strings"
	"time"

	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/service"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AnimeFilterSet represents a collection of filters to apply to the anime list
type AnimeFilterSet struct {
	statusFilters        []domain.MediaStatus // Empty slice means no status filter
	hasAvailableEpisodes bool                 // Filter to only anime with aired but unwatched episodes
	isFinishedAiring     bool                 // Filter to anime that have fully completed airing
}

// AnimeListModel handles displaying and interacting with the anime list
type AnimeListModel struct {
	config        *config.Config
	animeService  *service.AnimeService
	width, height int
	loading       bool
	loadError     error
	spinner       spinner.Model
	filters       AnimeFilterSet
	cursor        int
	allAnime      []*domain.Anime // All anime from the service
	filteredAnime []*domain.Anime // Anime after applying filters
}

// NewAnimeListModel creates a new anime list model
func NewAnimeListModel(cfg *config.Config, animeService *service.AnimeService) *AnimeListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	// Default filters - initially show only CURRENT anime
	defaultFilters := AnimeFilterSet{
		statusFilters: []domain.MediaStatus{domain.StatusCurrent},
	}

	return &AnimeListModel{
		config:        cfg,
		animeService:  animeService,
		loading:       true,
		spinner:       s,
		filters:       defaultFilters,
		cursor:        0,
		allAnime:      []*domain.Anime{},
		filteredAnime: []*domain.Anime{},
	}
}

// loadAnimeList loads the anime list from the service
func loadAnimeList(animeService *service.AnimeService) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := animeService.LoadAnimeList(ctx); err != nil {
			log.Error("Failed to load anime list", "error", err)
			return AnimeListErrorMsg{Error: err}
		}

		log.Info("Anime list loaded successfully.  Sending AnimeListLoadedMsg")
		return AnimeListLoadedMsg{}
	}
}

// Resize updates the model with new dimensions
func (m *AnimeListModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles messages and updates the model
func (m *AnimeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case "enter":
			// TODO: Implement view detail of selected anime
			log.Info("View anime detail", "title", m.getSelectedAnime().Title.English, "id", m.getSelectedAnime().ID)
		case "r":
			// Refresh anime list
			m.loading = true
			m.loadError = nil
			return m, loadAnimeList(m.animeService)
		}

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		return m, spinnerCmd

	case AnimeListLoadedMsg:
		log.Debug("Anime list loaded")
		m.loading = false
		m.allAnime = m.animeService.GetAnimeList()
		m.applyFilters()

	case AnimeListErrorMsg:
		log.Debug("Anime list load error", "error", msg.Error)
		m.loading = false
		m.loadError = msg.Error
	}

	if m.loading {
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		return m, spinnerCmd
	}

	return m, cmd
}

// toggleStatusFilter toggles a status filter based on the key pressed
func (m *AnimeListModel) toggleStatusFilter(key string) {
	var status domain.MediaStatus

	switch key {
	case "1":
		status = domain.StatusCurrent
	case "2":
		status = domain.StatusPlanning
	case "3":
		status = domain.StatusCompleted
	case "4":
		status = domain.StatusDropped
	case "5":
		status = domain.StatusPaused
	case "6":
		status = domain.StatusRepeating
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
		m.filters.statusFilters = append(m.filters.statusFilters[:index], m.filters.statusFilters[index+1:]...)

		// If we removed all filters, default to CURRENT
		if len(m.filters.statusFilters) == 0 {
			m.filters.statusFilters = []domain.MediaStatus{domain.StatusCurrent}
		}
	} else {
		// Status not in filters, add it
		m.filters.statusFilters = append(m.filters.statusFilters, status)
	}
}

// toggleHasNewEpisodesFilter toggles the new episodes filter
func (m *AnimeListModel) toggleHasNewEpisodesFilter() {
	m.filters.hasAvailableEpisodes = !m.filters.hasAvailableEpisodes
}

// toggleIsFinishedAiringFilter toggles the completed airing filter
func (m *AnimeListModel) toggleIsFinishedAiringFilter() {
	m.filters.isFinishedAiring = !m.filters.isFinishedAiring
}

// View renders the anime list model
func (m *AnimeListModel) View() string {
	if m.loading {
		return styles.CenteredView(
			m.width,
			m.height,
			fmt.Sprintf("%s Loading anime list...", m.spinner.View()),
		)
	}

	if m.loadError != nil {
		errorMsg := fmt.Sprintf("Error loading anime list: %v\n\nPress 'r' to retry.", m.loadError)
		return styles.CenteredView(
			m.width,
			m.height,
			styles.ContentBox(m.width-20, errorMsg, 1),
		)
	}

	// Build the view
	header := styles.Header(m.width, "Hisame - Anime List")
	filterStatus := m.renderFilterStatus()
	content := m.renderAnimeList()

	// Layout the components
	return fmt.Sprintf("%s\n\n%s\n\n%s", header, filterStatus, content)
}

// Init initializes the model
func (m *AnimeListModel) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		loadAnimeList(m.animeService),
	)
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
		statusMatch := false
		for _, status := range m.filters.statusFilters {
			if anime.UserData.Status == status {
				statusMatch = true
				break
			}
		}

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
			// Check if there are unwatched available episodes
			hasNewEps := anime.Episodes > 0 && anime.UserData.Progress < anime.Episodes
			if !hasNewEps {
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

// renderAnimeList renders the anime list for the current tab
func (m *AnimeListModel) renderAnimeList() string {
	animeList := m.filteredAnime

	if len(animeList) == 0 {
		return styles.CenteredText(m.width, "No anime found in this category")
	}

	// Calculate available height for the list
	availableHeight := m.height - 10 // Subtract space for header, tabs, and margins
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine visible range
	visibleCount := min(len(animeList), availableHeight-1) // Reserve space for header row

	// Adjust starting index to keep cursor in view
	startIdx := 0
	if m.cursor >= visibleCount {
		startIdx = m.cursor - visibleCount + 1
	}

	endIdx := startIdx + visibleCount
	if endIdx > len(animeList) {
		endIdx = len(animeList)
	}

	// Styles for list items
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(m.width-4).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Width(m.width-4).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Width(m.width-4).
		Padding(0, 1)

	// Build the list with header
	var listContent string

	// Add column headers
	headerText := fmt.Sprintf("%-3s %-50s %10s", "Sts", "Title", "Progress")
	listContent += headerStyle.Render(headerText) + "\n"
	//headerText := fmt.Sprintf("%-50s %-10s", "Title", "Progress")
	//listContent += headerStyle.Render(headerText) + "\n"

	// Add a separator line
	separatorLine := strings.Repeat("─", m.width-6) // Adjust width to fit inside the box
	listContent += separatorLine + "\n"

	// Add anime items
	for i := startIdx; i < endIdx; i++ {
		anime := animeList[i]

		itemText := m.formatAnimeListItem(anime)

		if i == m.cursor {
			listContent += selectedStyle.Render(itemText) + "\n"
		} else {
			listContent += normalStyle.Render(itemText) + "\n"
		}
	}

	// Add pagination indicator if needed
	if len(animeList) > visibleCount {
		pagination := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(animeList))
		listContent += styles.CenteredText(m.width-4, pagination)
	}

	return styles.ContentBox(m.width-2, listContent, 1)
}

// truncateString cuts a string to fit within maxWidth visual width
func truncateString(s string, maxWidth int) string {
	width := 0
	for i, r := range s {
		charWidth := runewidth.RuneWidth(r)
		// Check if adding this rune would exceed maxWidth
		if width+charWidth > maxWidth-3 { // Reserve space for "..."
			return s[:i] + "..."
		}
		width += charWidth
	}
	return s // Return as is if it fits
}

// formatAnimeListItem formats a single anime list item
func (m *AnimeListModel) formatAnimeListItem(anime *domain.Anime) string {
	title := anime.Title.Preferred(m.config.UI.TitleLanguage)

	titleWidth := 50
	truncatedTitle := truncateString(title, titleWidth)

	// Pad with spaces to ensure consistent width
	paddedTitle := truncatedTitle
	titleVisualWidth := 0
	for _, r := range truncatedTitle {
		titleVisualWidth += runewidth.RuneWidth(r)
	}
	if titleVisualWidth < titleWidth {
		paddedTitle = truncatedTitle + strings.Repeat(" ", titleWidth-titleVisualWidth)
	}

	progress := ""
	if anime.UserData != nil {
		if anime.Episodes > 0 {
			progress = fmt.Sprintf("%d/%d", anime.UserData.Progress, anime.Episodes)
		} else {
			progress = fmt.Sprintf("%d/?", anime.UserData.Progress)
		}
	}

	// Get status indicator
	statusIndicator := "[?]"
	if anime.UserData != nil {
		switch anime.UserData.Status {
		case domain.StatusCurrent:
			statusIndicator = "[W]"
		case domain.StatusPlanning:
			statusIndicator = "[P]"
		case domain.StatusCompleted:
			statusIndicator = "[C]"
		case domain.StatusDropped:
			statusIndicator = "[D]"
		case domain.StatusPaused:
			statusIndicator = "[H]"
		case domain.StatusRepeating:
			statusIndicator = "[R]"
		}
	}

	// Format with proper spacing
	return fmt.Sprintf("%s %s %10s", statusIndicator, paddedTitle, progress)
}

// getSelectedAnime returns the currently selected anime or nil if none
func (m *AnimeListModel) getSelectedAnime() *domain.Anime {
	animeList := m.filteredAnime
	if len(animeList) == 0 || m.cursor >= len(animeList) {
		return nil
	}
	return animeList[m.cursor]
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

	// Join all filter sections
	filterLine := " Status -> " + strings.Join(statusIndicators, " ") + " " + episodeFilters
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
