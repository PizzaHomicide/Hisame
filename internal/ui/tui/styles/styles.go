package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DEDEDE"))

	Instruction = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89CFF0"))

	Url = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#43BF6D")).
		Underline(true)
)

// Header creates a styled header
func Header(width int, title string) string {
	return Title.
		Width(width).
		Align(lipgloss.Center).
		Render(title)
}

// ContentBox creates a content area with optional padding
func ContentBox(width int, content string, padding int) string {
	return lipgloss.NewStyle().
		Width(width).
		Padding(padding).
		Render(content)
}

func CenteredView(width int, height int, content string) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func CenteredText(width int, text string) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(text)
}
