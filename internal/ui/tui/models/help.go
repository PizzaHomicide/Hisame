package models

import (
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModel represents a help "modal" that replaces the current view
type HelpModel struct {
	width, height int
}

// NewHelpModel creates a new help model
func NewHelpModel() *HelpModel {
	return &HelpModel{}
}

// Resize updates the dimensions of the help model
func (m *HelpModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// View returns the help content for the specified context
func (m *HelpModel) View(contextView View) string {
	// Get context-specific help content
	title, content := m.getHelpContent(contextView)

	// Create header using your predefined style
	header := styles.Header(m.width, title)

	// Format help content with section styles
	formattedContent := formatHelpContent(content)

	// Main content in a box
	contentWidth := m.width - 4 // Account for some padding
	contentBox := styles.ContentBox(contentWidth, formattedContent, 1)

	// Footer with exit instructions
	footer := styles.CenteredText(m.width, styles.Info.Render("Press ESC to return"))

	// Combine all elements for the view
	helpContent := lipgloss.JoinVertical(
		lipgloss.Left,
		"", // Top margin
		header,
		"", // Spacing
		contentBox,
		"", // Spacing
		footer,
	)

	// Center the entire view if needed
	return styles.CenteredView(m.width, m.height, helpContent)
}

// getHelpContent returns context-sensitive help title and content
func (m *HelpModel) getHelpContent(contextView View) (string, string) {
	var title string
	var content string

	// Global commands section - shown in all contexts
	globalCommands := `## Global Commands

* Ctrl+C: Quit application
* Ctrl+H: Toggle help
* Ctrl+L: Logout (clear token)
* ESC: Close modal/cancel current action

`

	// Context-specific help
	switch contextView {
	case ViewAuth:
		title = "Authentication Help"
		content = globalCommands + `## Auth View Commands

TODO - this is the auth view help
`

	case ViewAnimeList:
		title = "Anime List Help"
		content = globalCommands + `## Anime List Commands

* ↑/k: Move cursor up
* ↓/j: Move cursor down
* Enter: View details of selected anime
* r: Refresh anime list

## Filtering

* 1: Toggle Watching filter
* 2: Toggle Planning filter
* 3: Toggle Completed filter
* 4: Toggle Dropped filter
* 5: Toggle Paused filter
* 6: Toggle Repeating filter

## Status Indicators

* [W]: Watching
* [P]: Planning
* [C]: Completed
* [D]: Dropped
* [H]: Paused (On Hold)
* [R]: Repeating
`

	default:
		title = "Hisame Help"
		content = globalCommands
	}

	return title, content
}

// formatHelpContent applies styles to different parts of the help content
func formatHelpContent(content string) string {
	lines := strings.Split(content, "\n")
	styledLines := make([]string, 0, len(lines))

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	for _, line := range lines {
		if strings.HasPrefix(line, "##") {
			// Section header - use bold purple to match your theme
			header := strings.TrimPrefix(line, "## ")
			styledLines = append(styledLines, sectionStyle.Render(header))
		} else if strings.HasPrefix(line, "*") {
			// List items - keep as is but with a slight indent
			styledLines = append(styledLines, "  "+line)
		} else {
			// Regular content
			styledLines = append(styledLines, line)
		}
	}

	return strings.Join(styledLines, "\n")
}
