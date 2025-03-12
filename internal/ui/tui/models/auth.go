package models

import (
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

type AuthModel struct {
	width, height  int
	authInProgress bool
}

func NewAuthModel() *AuthModel {
	return &AuthModel{}
}

func (m *AuthModel) Init() tea.Cmd {
	return nil
}

func (m *AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			log.Info("Start login..")
			m.authInProgress = true
			// TODO: Login process
			return m, mockLogin()
		}
	}

	return m, nil
}

func (m *AuthModel) View() string {
	contentWidth := min(m.width, 120)

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
	content := styles.CenteredText(contentWidth-2,
		styles.Info.Render("You need to authenticate with AniList to use Hisame."))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-2,
		styles.Info.Render("When you press 'l' a browser will open to authenticate with Anilist")) + "\n"
	content += styles.CenteredText(contentWidth-2,
		styles.Info.Render("After seeing the Hisame login success screen in your browser, continue in this application")) + "\n\n"

	content += styles.CenteredText(contentWidth-2,
		styles.Info.Render("Press 'l' to login or 'ctrl+c' to quit."))

	return content
}

func (m *AuthModel) authInProgressContent(contentWidth int) string {
	content := styles.CenteredText(contentWidth-2, styles.Info.Render("Authenticating to AniList..."))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-2,
		styles.Info.Render("If your browser didn't open automatically, please visit the following URL:"))
	content += "\n\n"

	content += styles.CenteredText(contentWidth-2, styles.Url.Render("https://anilist.co/"))

	return content
}

// mockLogin 'simulates' an async login process by running it in a goroutine, waiting a few seconds, and completing
// It exists only to aid development of the TUI and is to be removed when the real login process is implemented.
func mockLogin() tea.Cmd {
	return func() tea.Msg {
		// Simulate delay
		log.Info("Authenticating to AniList...")
		time.Sleep(15 * time.Second)
		log.Info("Login successful.  Token received")

		return AuthCompletedMsg{
			Token: "foobar",
		}
	}
}
