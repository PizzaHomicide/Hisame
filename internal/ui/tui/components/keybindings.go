package components

import (
	"fmt"
	"strings"

	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// KeyBinding represents a single key and its description for the keybinding bar
type KeyBinding struct {
	Key  string
	Desc string
}

// keyStyle is used to highlight keyboard shortcuts in UI
var keyStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7D56F4")).
	Bold(true)

// KeyBindingsBar creates a styled footer showing a set of keybindings
// width: The width of the screen to center the bar
// bindings: The list of keybindings to display
func KeyBindingsBar(width int, bindings []KeyBinding) string {
	var parts []string
	for _, b := range bindings {
		parts = append(parts, fmt.Sprintf("%s: %s",
			keyStyle.Render(b.Key),
			b.Desc))
	}

	keyBar := styles.Info.Render(strings.Join(parts, " • "))
	return styles.CenteredText(width, keyBar)
}
