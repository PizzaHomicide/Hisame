package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Text styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DEDEDE"))

	Url = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#43BF6D")).
		Underline(true)

	FilterStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 2)
)

// Layout helpers
func Header(width int, title string) string {
	return Title.
		Width(width).
		Align(lipgloss.Center).
		Render(title)
}

func ContentBox(width int, content string, padding int) string {
	return lipgloss.NewStyle().
		Width(width).
		Padding(padding).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
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
