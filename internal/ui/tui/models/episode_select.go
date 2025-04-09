package models

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/log"
	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
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
func NewEpisodeSelectModel(episodes []player.AllAnimeEpisodeInfo, animeTitle string) *EpisodeSelectModel {
	input := textinput.New()
	input.Placeholder = "Filter episodes..."
	input.Width = 30
	input.SetValue("")

	hasMultiCours := false
	for _, ep := range episodes {
		if fmt.Sprintf("%d", ep.OverallEpisodeNumber) != ep.AllAnimeEpisodeNumber {
			hasMultiCours = true
			break
		}
	}

	return &EpisodeSelectModel{
		searchInput:    input,
		searchMode:     false,
		cursor:         0,
		episodes:       episodes,
		filtered:       episodes,
		animeTitle:     animeTitle,
		viewportOffset: 0,
		hasMultiCours:  hasMultiCours,
	}
}

func (m *EpisodeSelectModel) ViewType() View {
	return ViewEpisodeSelect
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
	return nil
}

// Update updates the model based on messages
func (m *EpisodeSelectModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in search mode, handle input differently
		if cmd := m.handleSearchModeKeyMsg(msg); cmd != nil {
			return m, cmd
		}

		if cmd := m.handleKeyMsg(msg); cmd != nil {
			return m, cmd
		}
	}

	return m, nil
}

func (m *EpisodeSelectModel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch kb.GetActionByKey(msg, kb.ContextEpisodeSelection) {
	case kb.ActionSelectEpisode:
		selectedEp := m.GetSelectedEpisode()
		if selectedEp != nil {
			return func() tea.Msg {
				return EpisodeMsg{
					Type:    EpisodeEventSelected,
					Episode: selectedEp,
				}
			}
		}
		log.Warn("Empty episode selected.  This should not be possible")
		return Handled("err:episode_select:empty_episode_selection")
	case kb.ActionEnableSearch:
		m.searchMode = true
		m.searchInput.Focus()
		return Handled("search:enable")
	case kb.ActionMoveDown:
		if len(m.filtered) > 0 && m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.ensureCursorVisible()
		}
		return Handled("cursor_move:down")
	case kb.ActionMoveUp:
		if m.cursor > 0 {
			m.cursor--
			m.ensureCursorVisible()
		}
		return Handled("cursor_move:up")
	case kb.ActionPageDown:
		pageSize := m.height - 11
		m.cursor += pageSize
		if m.cursor >= len(m.filtered) {
			m.cursor = len(m.filtered) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureCursorVisible()
		return Handled("cursor_move:pgdown")
	case kb.ActionPageUp:
		pageSize := m.height - 11
		m.cursor -= pageSize
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureCursorVisible()
		return Handled("cursor_move:pgup")
	}

	return nil
}

func (m *EpisodeSelectModel) handleSearchModeKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if !m.searchMode {
		return nil
	}
	switch kb.GetActionByKey(msg, kb.ContextSearchMode) {
	case kb.ActionBack:
		// Cancels search, clearing the filter
		m.searchMode = false
		m.searchInput.SetValue("")
		m.applyFilter()
		return Handled("search:exit")
	case kb.ActionSearchComplete:
		m.searchMode = false
		m.applyFilter()
		return Handled("search:apply")
	}

	// Let the text input model handle other keys
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Apply filters as we type
	m.applyFilter()

	return cmd
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
			fuzzy.Match(query, ep.AllAnimeName) ||
			fuzzy.Match(query, ep.PreferredTitle) {
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
	m.ensureCursorVisible()

}

// ensureCursorVisible adjusts the viewport offset to keep the cursor visible
func (m *EpisodeSelectModel) ensureCursorVisible() {
	// If no filtered episodes, reset cursor and offset
	if len(m.filtered) == 0 {
		m.cursor = 0
		m.viewportOffset = 0
		return
	}

	// Ensure cursor is within filtered episodes range
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}

	// Calculate available height for the list
	availableHeight := m.height - 10 // Subtract space for header, footer, and margins
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Adjust viewport to show as many entries as possible from the start
	// while keeping the cursor visible
	visibleCount := min(len(m.filtered), availableHeight-1)

	// If total filtered entries fit in viewport, reset offset
	if len(m.filtered) <= visibleCount {
		m.viewportOffset = 0
		return
	}

	// Ensure cursor is within current viewport
	if m.cursor < m.viewportOffset {
		// Cursor is above viewport, adjust offset
		m.viewportOffset = m.cursor
	}

	// Ensure cursor is within visible range from the bottom
	if m.cursor >= m.viewportOffset+visibleCount {
		// Cursor is below viewport, adjust offset to show last entries
		m.viewportOffset = max(0, m.cursor-visibleCount+1)
	}

	// Additional check to ensure we can fill the viewport if possible
	maxPossibleOffset := max(0, len(m.filtered)-visibleCount)
	if m.viewportOffset > maxPossibleOffset {
		m.viewportOffset = maxPossibleOffset
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
			"Ep #", "Cour #", "AllAnimeName", "Season", "Source")
	} else {
		headerText = fmt.Sprintf("%-5s %-70s %-20s %10s",
			"Ep #", "AllAnimeName", "Season", "Source")
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
	title := episode.AllAnimeName

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
