package models

import (
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/components"
	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a single item shown in the menu
type MenuItem struct {
	// Display text shown to the user
	Text string
	// Command executed when an item is selected
	Command tea.Cmd
}

type MenuModel struct {
	Title         string
	Items         []MenuItem
	Cursor        int
	width, height int
}

func (m *MenuModel) ViewType() View {
	return ViewMenu
}

func NewMenuModel(title string, items []MenuItem) *MenuModel {
	return &MenuModel{
		Title:  title,
		Items:  items,
		Cursor: 0,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch kb.GetActionByKey(msg, kb.ContextMenu) {
		case kb.ActionMoveUp:
			if m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil

		case kb.ActionMoveDown:
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
			return m, nil

		case kb.ActionSelectMenuItem:
			// Safety fallback, if no items just return nil cmd
			if len(m.Items) == 0 {
				return m, nil
			}

			selected := m.Items[m.Cursor]
			log.Info("Menu item selected", "title", m.Title, "item", selected.Text)
			return m, selected.Command
		}
	}

	return m, nil
}

func (m *MenuModel) View() string {
	if len(m.Items) == 0 {
		return styles.CenteredText(m.width, "No menu items available")
	}

	header := styles.Header(m.width, m.Title)

	cursorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Width(m.width-8). // Account for padding and cursor indicator
		Padding(0, 1)

	itemStyle := lipgloss.NewStyle().
		Width(m.width-8).
		Padding(0, 1)

	var menuContent string
	for i, item := range m.Items {
		var renderedItem string
		if i == m.Cursor {
			renderedItem = "> " + cursorStyle.Render(item.Text)
		} else {
			renderedItem = "  " + itemStyle.Render(item.Text)
		}
		menuContent += renderedItem + "\n"
	}

	content := styles.ContentBox(m.width-4, menuContent, 1)

	keyBindings := []components.KeyBinding{
		{"↑/↓", "Navigate"},
		{"Enter", "Select"},
		{"Esc", "Cancel"},
	}
	footer := components.KeyBindingsBar(m.width, keyBindings)

	// Combine all elements
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"", // Add an empty line for spacing
		content,
		"", // Add an empty line for spacing
		footer,
	)
}

func (m *MenuModel) Resize(width, height int) {
	m.width = width
	m.height = height
}
