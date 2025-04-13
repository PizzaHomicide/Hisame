package models

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/components"
	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/util"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// AnimeDetailsModel displays detailed information about a single anime
type AnimeDetailsModel struct {
	width, height int
	anime         *domain.Anime
	viewport      viewport.Model // For scrolling content
}

// NewAnimeDetailsModel creates a new anime details model
func NewAnimeDetailsModel(anime *domain.Anime) *AnimeDetailsModel {
	vp := viewport.New(80, 20) // Default size, will be updated in Resize()

	return &AnimeDetailsModel{
		anime:    anime,
		viewport: vp,
	}
}

func (m *AnimeDetailsModel) ViewType() View {
	return ViewAnimeDetails
}

// Init initializes the model
func (m *AnimeDetailsModel) Init() tea.Cmd {
	content := m.generateContent()
	m.viewport.SetContent(content)
	return nil
}

// Update handles messages
func (m *AnimeDetailsModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
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

	case tea.MouseMsg:
		// Handle mouse scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the anime details view
func (m *AnimeDetailsModel) View() string {
	// Generate header with anime title
	header := styles.Header(m.width, "Details: "+m.anime.Title.Preferred)

	// Viewport content (scrollable)
	viewportContent := m.viewport.View()

	// Define keybindings to be displayed in the footer
	keyBindings := []components.KeyBinding{
		{"↑/↓", "Scroll"},
		{"PgUp/PgDn", "Page scroll"},
		{"Ctrl+h", "Help"},
		{"Esc", "Return"},
	}
	footer := components.KeyBindingsBar(m.width, keyBindings)

	// Join all components with proper spacing
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"", // Add an empty line for spacing
		styles.ContentBox(m.width-2, viewportContent, 1),
		"", // Add an empty line for spacing
		footer,
	)
}

// Resize updates the dimensions of the model
func (m *AnimeDetailsModel) Resize(width, height int) {
	m.width = width
	m.height = height

	// Adjust viewport dimensions
	viewportWidth := width - 4    // Account for borders/padding
	viewportHeight := height - 10 // Account for header, footer, spacing

	// Ensure we don't set negative dimensions
	if viewportWidth < 1 {
		viewportWidth = 1
	}
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight

	// Regenerate content for the new width
	content := m.generateContent()
	m.viewport.SetContent(content)
}

// generateContent creates the detailed text content for the anime
func (m *AnimeDetailsModel) generateContent() string {
	anime := m.anime
	if anime == nil {
		return "Error: No anime data available"
	}

	// Determine content width (account for padding)
	contentWidth := m.width - 6
	if contentWidth < 60 {
		contentWidth = 60 // Minimum reasonable width
	}

	var b strings.Builder

	// Styles for different parts of the content
	sectionTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	fieldNameStyle := lipgloss.NewStyle().Bold(true)

	// Basic information section
	b.WriteString(sectionTitleStyle.Render("Anime Information"))
	b.WriteString("\n\n")

	// Format titles
	b.WriteString(fieldNameStyle.Render("Title (English): "))
	b.WriteString(anime.Title.English)
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Title (Romaji): "))
	b.WriteString(anime.Title.Romaji)
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Title (Native): "))
	b.WriteString(anime.Title.Native)
	b.WriteString("\n\n")

	// Format metadata
	b.WriteString(fieldNameStyle.Render("Format: "))
	b.WriteString(anime.Format)
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Status: "))
	b.WriteString(anime.Status)
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Episodes: "))
	if anime.Episodes > 0 {
		b.WriteString(fmt.Sprintf("%d", anime.Episodes))
	} else {
		b.WriteString("Unknown")
	}
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Season: "))
	if anime.Season != "" && anime.SeasonYear != "" {
		b.WriteString(fmt.Sprintf("%s %s", anime.Season, anime.SeasonYear))
	} else {
		b.WriteString("Unknown")
	}
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("Average Score: "))
	if anime.AverageScore > 0 {
		b.WriteString(fmt.Sprintf("%.1f", anime.AverageScore))
	} else {
		b.WriteString("Not rated")
	}
	b.WriteString("\n\n")

	// Next airing episode
	if anime.NextAiringEp != nil {
		b.WriteString(fieldNameStyle.Render("Next Episode: "))
		b.WriteString(fmt.Sprintf("Episode %d airing in %s",
			anime.NextAiringEp.Episode,
			strings.TrimSpace(util.FormatTimeUntilAiring(anime.NextAiringEp.TimeUntilAir))))
		b.WriteString("\n\n")
	}

	// User's personal information section
	if anime.UserData != nil {
		b.WriteString(sectionTitleStyle.Render("Your Information"))
		b.WriteString("\n\n")

		b.WriteString(fieldNameStyle.Render("Status: "))
		b.WriteString(string(anime.UserData.Status))
		b.WriteString("\n")

		b.WriteString(fieldNameStyle.Render("Progress: "))
		if anime.Episodes > 0 {
			b.WriteString(fmt.Sprintf("%d/%d episodes", anime.UserData.Progress, anime.Episodes))
		} else {
			b.WriteString(fmt.Sprintf("%d/? episodes", anime.UserData.Progress))
		}
		b.WriteString("\n")

		b.WriteString(fieldNameStyle.Render("Score: "))
		if anime.UserData.Score > 0 {
			b.WriteString(fmt.Sprintf("%.1f", anime.UserData.Score))
		} else {
			b.WriteString("Not rated")
		}
		b.WriteString("\n")

		if anime.UserData.StartDate != "" {
			b.WriteString(fieldNameStyle.Render("Started: "))
			b.WriteString(anime.UserData.StartDate)
			b.WriteString("\n")
		}

		if anime.UserData.EndDate != "" {
			b.WriteString(fieldNameStyle.Render("Completed: "))
			b.WriteString(anime.UserData.EndDate)
			b.WriteString("\n")
		}

		if anime.UserData.Notes != "" {
			b.WriteString("\n")
			b.WriteString(fieldNameStyle.Render("Notes:"))
			b.WriteString("\n")
			b.WriteString(anime.UserData.Notes)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Alternative titles section
	if len(anime.Synonyms) > 0 {
		b.WriteString(sectionTitleStyle.Render("Alternative Titles"))
		b.WriteString("\n\n")

		for _, synonym := range anime.Synonyms {
			b.WriteString("• ")
			b.WriteString(synonym)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}
