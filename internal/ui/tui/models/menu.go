package models

import (
	"strings"

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
	// IsSeparator indicates that this is a visual separator, not a selectable item
	IsSeparator bool
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
	m.ensureValidCursor()
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch kb.GetActionByKey(msg, kb.ContextMenu) {
		case kb.ActionMoveUp:
			m.moveCursorUp()
			return m, nil

		case kb.ActionMoveDown:
			m.moveCursorDown()
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

	var menuContent string
	for i, item := range m.Items {
		menuContent += item.Render(m.width, i == m.Cursor)
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

// Render renders a menu item with the given parameters
func (item MenuItem) Render(width int, isSelected bool) string {
	if item.IsSeparator {
		return item.renderSeparator(width)
	}
	return item.renderSelectable(width, isSelected)
}

// renderSeparator renders the item as a separator (not selectable)
func (item MenuItem) renderSeparator(width int) string {
	if item.Text == "" {
		// No separator text, so just use a plain line renderSeparator
		return "  " + strings.Repeat("-", width-10) + "\n"
	}
	// Calculate the space available for dashes
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	textWidth := lipgloss.Width(item.Text)
	availableWidth := width - 10                   // Account for margins and padding
	dashesNeeded := availableWidth - textWidth - 2 // -2 for spaces around text

	// If text is too long, just show the text
	if dashesNeeded <= 0 {
		return "  " + separatorStyle.Render(item.Text) + "\n"
	}

	// Create dashes on both sides of the text
	dashesPerSide := dashesNeeded / 2
	leftDashes := strings.Repeat("─", dashesPerSide)
	rightDashes := strings.Repeat("─", dashesNeeded-dashesPerSide)

	// Combine into the separator with text
	separator := leftDashes + " " + separatorStyle.Render(item.Text) + " " + rightDashes

	return "  " + separator + "\n"
}

// renderSelectable renders the item as a selectable menu item
func (item MenuItem) renderSelectable(width int, isSelected bool) string {
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Width(width-8).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Width(width-8).
		Padding(0, 1)

	// Determine style based on selection
	var renderedItem string
	if isSelected {
		renderedItem = selectedStyle.Render(item.Text)
	} else {
		renderedItem = normalStyle.Render(item.Text)
	}

	// Add cursor indicator
	if isSelected {
		renderedItem = "> " + renderedItem
	} else {
		renderedItem = "  " + renderedItem
	}

	return renderedItem + "\n"
}

// ensureValidCursor ensures the cursor is on a selectable item when the menu is first created
func (m *MenuModel) ensureValidCursor() {
	log.Trace("Ensuring valid cursor", "cursor", m.Cursor)
	if len(m.Items) == 0 {
		log.Trace("No item, early return")
		return
	}

	// If we're already on a non-separator, we're good
	if !m.Items[m.Cursor].IsSeparator {
		log.Trace("Already on a non-separator!", "item", m.Items[m.Cursor].Text)
		return
	}

	// moveCursorDown handles for separators, so this will move to the first non-separator if any
	log.Trace("Trying to move down")
	m.moveCursorDown()
}

// moveCursorUp moves the cursor up to the previous selectable item
func (m *MenuModel) moveCursorUp() {
	startPos := m.Cursor
	m.Cursor--

	// Keep moving up until we find a selectable item or hit the top
	for m.Cursor >= 0 && m.Items[m.Cursor].IsSeparator {
		m.Cursor--
	}

	// If we went past the top, restore the original position
	if m.Cursor < 0 {
		m.Cursor = startPos
	}
}

// moveCursorDown moves the cursor down to the next selectable item
func (m *MenuModel) moveCursorDown() {
	startPos := m.Cursor
	m.Cursor++

	// Keep moving down until we find a selectable item or hit the bottom
	for m.Cursor < len(m.Items) && m.Items[m.Cursor].IsSeparator {
		m.Cursor++
	}

	// If we went past the bottom, restore the original position
	if m.Cursor >= len(m.Items) {
		m.Cursor = startPos
	}
}
