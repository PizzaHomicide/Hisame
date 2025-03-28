package models

import (
	"fmt"
	"strings"

	"github.com/PizzaHomicide/hisame/internal/player"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mattn/go-runewidth"
)

// EpisodeSelectModel represents the episode selection modal
type EpisodeSelectModel struct {
	width, height  int
	episodes       []player.AllAnimeEpisodeInfo
	filtered       []player.AllAnimeEpisodeInfo
	cursor         int
	searchInput    textinput.Model
	searchMode     bool
	animeTitle     string
	hasMultiCours  bool // Flag to indicate if we need to show cour episode numbers
	viewportOffset int  // For scrolling
}

// NewEpisodeSelectModel creates a new episode selection modal
func NewEpisodeSelectModel() *EpisodeSelectModel {
	input := textinput.New()
	input.Placeholder = "Filter episodes..."
	input.Width = 30

	return &EpisodeSelectModel{
		searchInput: input,
		searchMode:  false,
		cursor:      0,
	}
}

func (m *EpisodeSelectModel) ViewType() View {
	return ViewEpisodeSelect
}

// SetEpisodes sets the episodes to display
func (m *EpisodeSelectModel) SetEpisodes(episodes []player.AllAnimeEpisodeInfo, animeTitle string) {
	m.episodes = episodes
	m.filtered = episodes
	m.animeTitle = animeTitle
	m.cursor = 0
	m.viewportOffset = 0
	m.searchInput.SetValue("")

	// Determine if we have multiple cours by checking if any show has different AllAnimeEpisodeNumber from OverallEpisodeNumber
	m.hasMultiCours = false
	for _, ep := range episodes {
		if fmt.Sprintf("%d", ep.OverallEpisodeNumber) != ep.AllAnimeEpisodeNumber {
			m.hasMultiCours = true
			break
		}
	}
}

// GetSelectedEpisode returns the currently selected episode
func (m *EpisodeSelectModel) GetSelectedEpisode() *player.AllAnimeEpisodeInfo {
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &m.filtered[m.cursor]
}

// Init initializes the model
func (m *EpisodeSelectModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update updates the model based on messages
func (m *EpisodeSelectModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// If in search mode, handle input differently
		if m.searchMode {
			switch msg.String() {
			case "esc":
				// Cancel search
				m.searchInput.SetValue("")
				m.searchMode = false
				m.applyFilter()
				return m, nil
			case "enter":
				// Apply search and exit search mode
				m.searchMode = false
				m.applyFilter()
				return m, nil
			}

			// Let the text input handle other keys
			m.searchInput, cmd = m.searchInput.Update(msg)

			// Apply filter as we type
			m.applyFilter()
			return m, cmd
		}

		// Normal mode key handling
		switch msg.String() {
		case "esc":
			// This will be handled by the parent to close the modal
			return m, nil

		case "enter":
			// Select the current episode
			selectedEp := m.GetSelectedEpisode()
			if selectedEp != nil {
				return m, func() tea.Msg {
					return EpisodeMsg{
						Type:    EpisodeEventSelected,
						Episode: selectedEp,
					}
				}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}

		case "down", "j":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}

		case "page-up":
			// Move up by a page
			pageSize := m.height - 10 // Approximate
			m.cursor -= pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureCursorVisible()

		case "page-down":
			// Move down by a page
			pageSize := m.height - 10 // Approximate
			m.cursor += pageSize
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureCursorVisible()

		case "ctrl+f":
			// Enter search mode
			m.searchMode = true
			m.searchInput.Focus()
			return m, textinput.Blink
		}
	}

	return m, cmd
}

// applyFilter filters episodes based on search input
func (m *EpisodeSelectModel) applyFilter() {
	query := m.searchInput.Value()
	if query == "" {
		m.filtered = m.episodes
		return
	}

	var filtered []player.AllAnimeEpisodeInfo
	for _, ep := range m.episodes {
		// Convert overall episode number to string for matching
		epNumStr := fmt.Sprintf("%d", ep.OverallEpisodeNumber)

		// Try fuzzy matching on episode numbers and title
		if fuzzy.Match(query, epNumStr) ||
			fuzzy.Match(query, ep.AllAnimeEpisodeNumber) ||
			fuzzy.Match(query, ep.Title) ||
			fuzzy.Match(query, ep.EnglishTitle) {
			filtered = append(filtered, ep)
		}
	}

	m.filtered = filtered

	// Reset cursor if needed
	if len(m.filtered) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
}

// ensureCursorVisible adjusts the viewport offset to keep the cursor visible
func (m *EpisodeSelectModel) ensureCursorVisible() {
	availableHeight := m.height - 10 // Adjust for header and footer

	// If cursor is above viewport, scroll up
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}

	// If cursor is below viewport, scroll down
	if m.cursor >= m.viewportOffset+availableHeight {
		m.viewportOffset = m.cursor - availableHeight + 1
	}
}

// View renders the episode selection modal
func (m *EpisodeSelectModel) View() string {
	// Build the view
	header := styles.Header(m.width, "Episode Selection - "+m.animeTitle)
	content := m.renderEpisodeList()

	if m.searchMode {
		// Show search input at the top of the content
		searchPrompt := styles.Title.Render("Search: ") + m.searchInput.View()
		content = lipgloss.JoinVertical(lipgloss.Left, searchPrompt, content)
	}

	// Show key bindings at the bottom
	keyBindings := " ↑/↓: Navigate • Enter: Select • Ctrl+f: Search • Esc: Cancel "
	footer := styles.FilterStatus.Render(keyBindings)

	// Layout the components
	return fmt.Sprintf("%s\n\n%s\n\n%s", header, content, footer)
}

// Resize updates the dimensions of the help model
func (m *EpisodeSelectModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// renderEpisodeList renders the list of episodes
func (m *EpisodeSelectModel) renderEpisodeList() string {
	if len(m.filtered) == 0 {
		if m.searchInput.Value() != "" {
			return styles.CenteredText(m.width, "No episodes match your filter")
		}
		return styles.CenteredText(m.width, "No episodes found")
	}

	// Calculate available height for the list
	availableHeight := m.height - 10 // Subtract space for header, footer, and margins
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine visible range
	visibleCount := min(len(m.filtered), availableHeight-1) // Reserve space for header row

	// Calculate the range of episodes to display
	startIdx := m.viewportOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}

	// Styles for list items
	//TODO:  Use styles package
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
	var headerText string
	if m.hasMultiCours {
		headerText = fmt.Sprintf("%-5s %-6s %-50s %-20s %10s",
			"Ep #", "Cour #", "Title", "Season", "Source")
	} else {
		headerText = fmt.Sprintf("%-5s %-70s %-20s %10s",
			"Ep #", "Title", "Season", "Source")
	}
	listContent += headerStyle.Render(headerText) + "\n"

	// Add a separator line
	separatorLine := strings.Repeat("─", m.width-6) // Adjust width to fit inside the box
	listContent += separatorLine + "\n"

	// Add episode items
	for i := startIdx; i < endIdx; i++ {
		episode := m.filtered[i]
		itemText := m.formatEpisodeListItem(episode)

		if i == m.cursor {
			listContent += selectedStyle.Render(itemText) + "\n"
		} else {
			listContent += normalStyle.Render(itemText) + "\n"
		}
	}

	// Add pagination indicator if needed
	if len(m.filtered) > visibleCount {
		pagination := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(m.filtered))
		listContent += styles.CenteredText(m.width-4, pagination)
	}

	return styles.ContentBox(m.width-2, listContent, 1)
}

// formatEpisodeListItem formats a single episode list item
func (m *EpisodeSelectModel) formatEpisodeListItem(episode player.AllAnimeEpisodeInfo) string {
	// Format episode number
	epNum := fmt.Sprintf("%d", episode.OverallEpisodeNumber)

	// Get title and truncate it
	title := episode.Title

	// Format season information
	season := fmt.Sprintf("%s %d", episode.Season, episode.Year)

	// Format based on whether we're showing cour numbers
	var result string
	if m.hasMultiCours {
		// Truncate title to fit
		truncatedTitle := util.TruncateString(title, 49)
		titleVisualWidth := runewidth.StringWidth(truncatedTitle)
		paddedTitle := truncatedTitle + strings.Repeat(" ", 49-titleVisualWidth)

		result = fmt.Sprintf("%-5s %-6s %-50s %-20s",
			epNum,
			episode.AllAnimeEpisodeNumber,
			paddedTitle,
			season)
	} else {
		// Truncate title to fit
		truncatedTitle := util.TruncateString(title, 69)
		titleVisualWidth := runewidth.StringWidth(truncatedTitle)
		paddedTitle := truncatedTitle + strings.Repeat(" ", 69-titleVisualWidth)

		result = fmt.Sprintf("%-5s %-70s %-20s",
			epNum,
			paddedTitle,
			season)
	}

	return result
}
