package models

import (
	"fmt"
	"strings"
	"unicode/utf8"

	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel displays contextual help with scrolling
type HelpModel struct {
	width, height int
	context       View
	viewport      viewport.Model
}

// NewHelpModel creates a new help model for the given context
func NewHelpModel(context View) *HelpModel {
	return &HelpModel{
		context:  context,
		viewport: viewport.New(0, 0),
	}
}

func (m *HelpModel) ViewType() View {
	return ViewHelp
}

// Init initializes the model
func (m *HelpModel) Init() tea.Cmd {
	// Set initial content if dimensions are available
	if m.width > 0 && m.height > 0 {
		m.updateContent()
	}
	return nil
}

// Update handles messages
func (m *HelpModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch kb.GetActionByKey(msg, kb.ContextHelp) {
		case kb.ActionMoveUp, kb.ActionMoveDown, kb.ActionPageUp, kb.ActionPageDown:
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		case kb.ActionMoveTop:
			m.viewport.GotoTop()
			return m, cmd
		case kb.ActionMoveBottom:
			m.viewport.GotoBottom()
			return m, cmd
		}

	}
	return m, cmd
}

// Resize updates the dimensions
func (m *HelpModel) Resize(width, height int) {
	m.width = width
	m.height = height

	// Update viewport dimensions
	contentWidth := width - 4    // Account for borders
	contentHeight := height - 10 // Account for header, footer, spacing

	// Ensure we don't set negative dimensions
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	m.viewport.Width = contentWidth
	m.viewport.Height = contentHeight

	// Update content for new dimensions
	m.updateContent()
}

// updateContent generates help content and updates the viewport
func (m *HelpModel) updateContent() {
	content := m.generateHelpContent()
	m.viewport.SetContent(content)
	// Reset to top when content changes
	m.viewport.GotoTop()
}

// View renders the help screen
func (m *HelpModel) View() string {
	title := m.getContextTitle()

	// Create header
	header := styles.Header(m.width, "Help: "+title)

	// Main content area with viewport
	contentView := m.viewport.View()

	// Footer with navigation help
	scrollText := "↑/↓: Scroll • PgUp/PgDn: Page scroll • Home/End: Goto top/bottom • ESC: Return"
	footer := styles.CenteredText(m.width, styles.Info.Render(scrollText))

	// Combine elements
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"", // Spacing
		styles.ContentBox(m.width-2, contentView, 1),
		"", // Spacing
		footer,
	)
}

// getContextTitle returns a user-friendly title for the context
func (m *HelpModel) getContextTitle() string {
	switch m.context {
	case ViewAuth:
		return "Authentication"
	case ViewAnimeList:
		return "Anime List"
	case ViewEpisodeSelect:
		return "Episode Selection"
	default:
		return "General"
	}
}

// formatKeybindingSection formats a section of keybindings with aligned colons
func (m *HelpModel) formatKeybindingSection(title string, bindings []kb.Binding, skipActions map[kb.Action]bool) string {
	if len(bindings) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title))
	b.WriteString("\n\n")

	// First pass: determine the maximum key width for alignment
	maxKeyWidth := 0
	for _, binding := range bindings {
		if skipActions != nil && skipActions[binding.Action] {
			continue
		}

		keyText := binding.KeyMap.Primary
		if binding.KeyMap.Secondary != "" {
			keyText += " or " + binding.KeyMap.Secondary
		}

		if width := utf8.RuneCountInString(keyText); width > maxKeyWidth {
			maxKeyWidth = width
		}
	}

	// Second pass: format each binding with aligned colons
	for _, binding := range bindings {
		if skipActions != nil && skipActions[binding.Action] {
			continue
		}

		keyText := binding.KeyMap.Primary
		if binding.KeyMap.Secondary != "" {
			keyText += " or " + binding.KeyMap.Secondary
		}

		// Create padding for alignment
		padding := strings.Repeat(" ", maxKeyWidth-utf8.RuneCountInString(keyText))

		b.WriteString(fmt.Sprintf("• %s%s : %s\n",
			lipgloss.NewStyle().Bold(true).Render(keyText),
			padding,
			binding.KeyMap.Help))
	}

	return b.String()
}

// generateHelpContent builds the complete help content
func (m *HelpModel) generateHelpContent() string {
	var b strings.Builder

	// Title style for sections
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	// Add context description section
	b.WriteString(titleStyle.Render(m.getContextTitle()))
	b.WriteString("\n\n")
	b.WriteString(m.getContextDescription())
	b.WriteString("\n\n")

	// Add keybindings section
	b.WriteString(titleStyle.Render("Keybindings"))
	b.WriteString("\n\n")

	// Global keybindings
	globalBindings := m.formatKeybindingSection("Global commands:", kb.ContextBindings[kb.ContextGlobal], nil)
	b.WriteString(globalBindings)

	// Build a map of global actions to avoid duplicating them in context-specific bindings
	globalActions := make(map[kb.Action]bool)
	for _, binding := range kb.ContextBindings[kb.ContextGlobal] {
		globalActions[binding.Action] = true
	}

	// Context-specific keybindings
	var contextName kb.ContextName

	switch m.context {
	case ViewAuth:
		contextName = kb.ContextAuth
	case ViewAnimeList:
		contextName = kb.ContextAnimeList
	case ViewEpisodeSelect:
		contextName = kb.ContextEpisodeSelection
	}

	if contextName != "" {
		// Add spacing between sections
		if globalBindings != "" {
			b.WriteString("\n")
		}

		sectionTitle := fmt.Sprintf("%s commands:", m.getContextTitle())
		contextBindings := m.formatKeybindingSection(sectionTitle, kb.ContextBindings[contextName], globalActions)
		b.WriteString(contextBindings)

		// Add filter details for AnimeList view
		if contextName == kb.ContextAnimeList {
			b.WriteString("\n")
			b.WriteString(m.getFilterDetails())
		}
	}

	// Search mode keybindings if applicable
	if m.context == ViewAnimeList || m.context == ViewEpisodeSelect {
		b.WriteString("\n")
		searchBindings := m.formatKeybindingSection("When in search mode:", kb.ContextBindings[kb.ContextSearchMode], nil)
		b.WriteString(searchBindings)
	}

	return b.String()
}

// getFilterDetails returns detailed explanation of filters for the anime list view
func (m *HelpModel) getFilterDetails() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	b.WriteString(titleStyle.Render("Filters"))
	b.WriteString("\n\n")

	b.WriteString("Status filters:\n\n")
	b.WriteString("• [W] : Watching - Shows anime you're currently watching\n")
	b.WriteString("• [P] : Planning - Shows anime you plan to watch in the future\n")
	b.WriteString("• [C] : Completed - Shows anime you've finished watching\n")
	b.WriteString("• [D] : Dropped - Shows anime you've stopped watching\n")
	b.WriteString("• [H] : On-Hold - Shows anime you've paused watching\n")
	b.WriteString("• [R] : Repeating - Shows anime you're rewatching\n\n")

	b.WriteString("Episode filters:\n\n")
	b.WriteString("• [A] : Available Episodes - Shows only anime with unwatched aired episodes\n")
	b.WriteString("• [F] : Finished Airing - Shows only anime that have completed their broadcast run\n\n")

	b.WriteString("Multiple filters can be active at once. Toggle each filter by pressing its corresponding key.\n")
	b.WriteString("If no status filters are active, the 'Watching' filter will be applied by default.\n")

	return b.String()
}

// getContextDescription returns help text for the current context
func (m *HelpModel) getContextDescription() string {
	switch m.context {
	case ViewAuth:
		return "The authentication screen allows you to connect Hisame with your AniList account.\n\n" +
			"When you press the login key, a browser window will open where you can authorize the application. " +
			"After completing authorization in your browser, you'll automatically return to Hisame."

	case ViewAnimeList:
		return "The anime list screen displays your AniList collection with filtering options.\n\n" +
			"Each anime entry shows information including progress, format, score, status, and upcoming episodes. " +
			"The '+' symbol indicates an anime has unwatched episodes available.\n\n" +
			"You can filter by status categories (watching, planning, etc.), search by title, " +
			"and directly play the next episode of a selected anime."

	case ViewEpisodeSelect:
		return "The episode selection screen allows you to choose a specific episode to watch.\n\n" +
			"Browse through available episodes, select one, and press Enter to begin playback. " +
			"You can use the search feature to quickly find specific episodes by number or title."

	default:
		return "Welcome to Hisame, a terminal UI for managing your AniList and watching anime."
	}
}
