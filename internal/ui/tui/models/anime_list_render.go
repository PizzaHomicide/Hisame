package models

// anime_list_render.go is responsible for visual representation of the anime list.
// It contains the rendering logic for the list view, including formatting individual
// anime entries, handling pagination, and proper display of anime metadata.

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"strings"
)

// renderAnimeList renders the anime list for the current filters
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
	headerText := fmt.Sprintf("%1s %-100s %8s %8s %5s %9s %5s %12s",
		" ", "Title", "Progress", "Format", "Score", "Status", "Next #", "Airing In")
	listContent += headerStyle.Render(headerText) + "\n"

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

// formatAnimeListItem formats a single anime list item for display
func (m *AnimeListModel) formatAnimeListItem(anime *domain.Anime) string {
	available := " " // Default: empty/space
	if anime.HasUnwatchedEpisodes() {
		available = "+"
	}

	title := anime.Title.Preferred

	// Truncate title to fit available space
	titleWidth := 100
	truncatedTitle := util.TruncateString(title, titleWidth)

	// Pad with spaces to ensure consistent width
	paddedTitle := truncatedTitle
	titleVisualWidth := 0
	for _, r := range truncatedTitle {
		titleVisualWidth += runewidth.RuneWidth(r)
	}
	if titleVisualWidth < titleWidth {
		paddedTitle = truncatedTitle + strings.Repeat(" ", titleWidth-titleVisualWidth)
	}

	// Format - TV, Movie, OVA, etc.
	format := "?"
	if anime.Format != "" {
		format = string(anime.Format)
	}

	// Progress
	progress := ""
	if anime.UserData != nil {
		if anime.Episodes > 0 {
			progress = fmt.Sprintf("%d/%d", anime.UserData.Progress, anime.Episodes)
		} else {
			progress = fmt.Sprintf("%d/?", anime.UserData.Progress)
		}
	}

	// Mean Score from AniList
	meanScore := "-"
	if anime.AverageScore > 0 {
		meanScore = fmt.Sprintf("%.0f", anime.AverageScore)
	}

	// Next episode number
	nextEpNum := ""
	if anime.NextAiringEp != nil {
		nextEpNum = fmt.Sprintf("%d", anime.NextAiringEp.Episode)
	}

	// Airing countdown
	airingIn := ""
	if anime.NextAiringEp != nil {
		airingIn = util.FormatTimeUntilAiring(anime.NextAiringEp.TimeUntilAir)
	} else if anime.Status == "FINISHED" {
		airingIn = "Finished"
	}

	// Status indicator
	statusText := "Unknown"
	if anime.UserData != nil {
		switch anime.UserData.Status {
		case domain.StatusCurrent:
			statusText = "Watching"
		case domain.StatusPlanning:
			statusText = "Planning"
		case domain.StatusCompleted:
			statusText = "Completed"
		case domain.StatusDropped:
			statusText = "Dropped"
		case domain.StatusPaused:
			statusText = "Paused"
		case domain.StatusRepeating:
			statusText = "Repeating"
		}
	}

	// Final formatted string
	return fmt.Sprintf("%s %-40s %8s %8s %5s %9s %5s %12s",
		available,
		paddedTitle,
		progress,
		format,
		meanScore,
		statusText,
		nextEpNum,
		airingIn)
}
