package models

import (
	"strings"
	"time"

	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingModel displays a loading indicator with contextual messages
type LoadingModel struct {
	width, height int
	title         string // Optional title for the loading box
	message       string // Primary message displayed with the spinner
	contextInfo   string // Optional additional context
	actionText    string // Optional action text/instruction
	spinner       spinner.Model
	startTime     time.Time // Track when loading started
}

// NewLoadingModel creates a new loading model with the required message
func NewLoadingModel(message string) *LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return &LoadingModel{
		message:   message,
		spinner:   s,
		startTime: time.Now(),
	}
}

// WithTitle adds an optional title to the loading box
func (m *LoadingModel) WithTitle(title string) *LoadingModel {
	m.title = title
	return m
}

// WithContextInfo adds additional context information
func (m *LoadingModel) WithContextInfo(info string) *LoadingModel {
	m.contextInfo = info
	return m
}

// WithActionText adds text describing a possible user action
func (m *LoadingModel) WithActionText(text string) *LoadingModel {
	m.actionText = text
	return m
}

// ViewType returns the type of view
func (m *LoadingModel) ViewType() View {
	return ViewLoading
}

// Init initializes the model
func (m *LoadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m *LoadingModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	log.Warn("Loading model received message it can't handle", "message", msg)
	return m, nil
}

// View renders the loading state
func (m *LoadingModel) View() string {
	// Calculate optimal content width - not too wide, not too narrow
	contentWidth := min(m.width-20, 80)
	if contentWidth < 40 {
		contentWidth = min(m.width-4, 40) // Ensure minimum reasonable width
	}

	// Special spinner style with more emphasis
	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9D86FF")).
		Bold(true).
		PaddingRight(1)

	// Message style for the primary message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	// Center alignment style for all content
	centerStyle := lipgloss.NewStyle().
		Width(contentWidth - 6). // Account for padding
		Align(lipgloss.Center)

	// Start building the content
	var contentBuilder strings.Builder

	// Add centered spinner and message
	primaryRow := spinnerStyle.Render(m.spinner.View()) + " " + messageStyle.Render(m.message)
	contentBuilder.WriteString(centerStyle.Render(primaryRow))

	// Add spacing and context info if present
	if m.contextInfo != "" {
		contextStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Italic(true).
			Width(contentWidth - 6).
			Align(lipgloss.Center)

		contentBuilder.WriteString("\n\n")
		contentBuilder.WriteString(contextStyle.Render(m.contextInfo))
	}

	// Add action text if present with distinctive styling
	if m.actionText != "" {
		actionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43BF6D")).
			Bold(true).
			Width(contentWidth-6).
			Align(lipgloss.Center).
			Padding(1, 0)

		contentBuilder.WriteString("\n\n")
		contentBuilder.WriteString(actionStyle.Render(m.actionText))
	}

	// Get the fully built content
	content := contentBuilder.String()

	// Create a bordered box with enhanced styling
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9D86FF")).
		Padding(2, 3).
		Width(contentWidth)

	// Create the final view
	var finalView string
	if m.title != "" {
		// If we have a title, use it in the header with special styling for emphasis
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			Align(lipgloss.Center).
			Width(contentWidth)

		header := titleStyle.Render(m.title)
		contentBox := boxStyle.Render(content)

		// Join with proper spacing
		finalView = lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			contentBox,
		)
	} else {
		// Otherwise just show the content box with a bit more padding
		finalView = boxStyle.Render(content)
	}

	return styles.CenteredView(m.width, m.height, finalView)
}

// Resize updates the dimensions of the loading model
func (m *LoadingModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// GetElapsedTime returns the time elapsed since loading started
func (m *LoadingModel) GetElapsedTime() time.Duration {
	return time.Since(m.startTime)
}
