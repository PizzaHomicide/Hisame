package models

import (
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the main application model that coordinates all child models.  It is the high level wrapper.
type AppModel struct {
	config        *config.Config
	activeView    View
	width, height int

	// Models used for various views
	authModel *AuthModel
}

// NewAppModel creates a new instance of the main application model
func NewAppModel(cfg *config.Config) AppModel {
	var initialView View

	// TODO: Validation on the token.
	if cfg.Auth.Token != "" {
		// Skip the auth view
		initialView = ViewAnimeList
	} else {
		initialView = ViewAuth
	}
	return AppModel{
		config:     cfg,
		activeView: initialView,
		authModel:  NewAuthModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	log.Info("Initialising Hisame TUI")

	// If starting application on anime list view, load the anime now
	if m.activeView == ViewAnimeList {
		log.Debug("TODO:  Load anime list")
		return nil
	}

	return nil
}

// Update handles messages and updates the models as appropriate
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			log.Info("Quit command received.  Shutting down...")
			return m, tea.Quit
		case "ctrl+h":
			log.Debug("Opening help panel")
			// TODO: Help panel
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		log.Debug("Window size changed", "width", m.width, "height", m.height)
	}

	// Delegate message processing to the active view
	switch m.activeView {
	case ViewAuth:
		return m.updateAuthView(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.activeView {
	case ViewAuth:
		return m.authModel.View()
	default:
		return "Unknown view\nPress ctrl+c to quit."
	}
}

// updateAuthView delegates message processing to
func (m AppModel) updateAuthView(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Any messages that require orchestration/view changing specific to the auth view
	switch msg.(type) {
	case AuthCompletedMsg:
		log.Info("Authentication successful")
		m.activeView = ViewAnimeList
		// TODO: Initialise/load data from AniList
		return m, nil
	}

	// Delegate other messages to the model
	authModel, cmd := m.authModel.Update(msg)
	m.authModel = authModel.(*AuthModel)

	return m, cmd
}
