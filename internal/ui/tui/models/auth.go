package models

import (
	"github.com/PizzaHomicide/hisame/internal/auth"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// HorizontalPadding - padding on each side of content
	HorizontalPadding = 2

	// MinWidth Minimum usable width for the UI
	MinWidth = 35

	// MinHeight Minimum usable height for the UI
	MinHeight = 12

	// MaxContentWidth Maximum content width
	MaxContentWidth = 120
)

type AuthModel struct {
	width, height  int
	authInProgress bool
	authUrl        string
}

func NewAuthModel() *AuthModel {
	return &AuthModel{
		authUrl: "Authentication URL not available",
	}
}

func (m *AuthModel) Init() tea.Cmd {
	return nil
}

func (m *AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			log.Info("Start login..")
			m.authInProgress = true
			return m, m.startAuth()
		}
	}

	return m, nil
}

// startAuth begins the authentication process
func (m *AuthModel) startAuth() tea.Cmd {
	authManager := auth.NewAuth()
	m.authUrl = authManager.LoginURL.String()
	return func() tea.Msg {
		result := authManager.DoAuth()
		m.authInProgress = false

		if result.Error != nil {
			return AuthMsg{
				Success: false,
				Error:   result.Error.Error(),
			}
		}

		return AuthMsg{
			Success: true,
			Token:   result.Token,
		}
	}
}

// Reset resets the auth model so it is ready to do a fresh login if necessary
func (m *AuthModel) Reset() {
	m.authInProgress = false
	m.authUrl = ""
}

func (m *AuthModel) View() string {
	// Reserve a few pixels for borders/padding
	availableWidth := m.width - HorizontalPadding

	contentWidth := min(availableWidth, MaxContentWidth)

	// If terminal is extremely small, show a simplified view
	if m.width < MinWidth || m.height < MinHeight {
		return "Terminal too small\nResize or press ctrl+c"
	}

	header := styles.Header(contentWidth, "Hisame")

	var content string
	if m.authInProgress {
		content = m.authInProgressContent(contentWidth)
	} else {
		content = m.initialContent(contentWidth)
	}

	// Box the content
	mainContent := styles.ContentBox(contentWidth, content, 1)

	// Join header and content
	combinedContent := lipgloss.JoinVertical(lipgloss.Center, header, mainContent)

	// Center everything in the terminal
	return styles.CenteredView(m.width, m.height, combinedContent)
}

func (m *AuthModel) initialContent(contentWidth int) string {
	content := styles.CenteredText(contentWidth-HorizontalPadding,
		styles.Info.Render("You need to authenticate with AniList to use Hisame."))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-HorizontalPadding,
		styles.Info.Render("When you press 'l' a browser will open to authenticate with Anilist")) + "\n"
	content += styles.CenteredText(contentWidth-HorizontalPadding,
		styles.Info.Render("After seeing the Hisame login success screen in your browser, continue in this application")) + "\n\n"

	content += styles.CenteredText(contentWidth-HorizontalPadding,
		styles.Info.Render("Press 'l' to login or 'ctrl+c' to quit."))

	return content
}

func (m *AuthModel) authInProgressContent(contentWidth int) string {
	content := styles.CenteredText(contentWidth-HorizontalPadding, styles.Info.Render("Authenticating to AniList..."))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-HorizontalPadding,
		styles.Info.Render("If your browser didn't open automatically, please visit the following URL:"))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-HorizontalPadding, styles.Url.Render(m.authUrl))

	return content
}

// Resize updates the dimensions of the auth model
func (m *AuthModel) Resize(width, height int) {
	m.width = width
	m.height = height
}
