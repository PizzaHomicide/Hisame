package models

import (
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/repository/anilist"
	"github.com/PizzaHomicide/hisame/internal/service"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

// AppModel is the main application model that coordinates all child models.  It is the high level wrapper.
type AppModel struct {
	config        *config.Config
	activeView    View  // Track the current active 'main view'
	activeModal   Modal // Track the current active 'modal overlay' if any
	width, height int

	// Models used for various views
	authModel *AuthModel
	helpModel *HelpModel

	// Services used for fetching and updating state
	animeService *service.AnimeService
}

// NewAppModel creates a new instance of the main application model
func NewAppModel(cfg *config.Config) AppModel {
	var initialView View
	var animeService *service.AnimeService

	if cfg.Auth.Token != "" {
		log.Info("Token found in config file.  Testing it to see if still valid")
		client, err := anilist.NewClient(cfg.Auth.Token)
		if err != nil {
			log.Warn("Failed to create anilist client with token from config.  Reauthentication required")
			initialView = ViewAuth
		} else {
			// Client initialised correct, so we can bypass auth.
			animeRepo := anilist.NewAnimeRepository(client)
			animeService = service.NewAnimeService(animeRepo)
			initialView = ViewAnimeList

		}
	} else {
		initialView = ViewAuth
	}
	return AppModel{
		config:       cfg,
		activeView:   initialView,
		activeModal:  ModalNone,
		authModel:    NewAuthModel(),
		helpModel:    NewHelpModel(),
		animeService: animeService,
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
		case "ctrl+l":
			log.Info("Logging out.  Cleaning up token from config file...")
			m.config.Auth.Token = ""
			config.UpdateConfig(func(conf *config.Config) {
				conf.Auth.Token = ""
			})
			// Throw back to login screen
			m.authModel.Reset()
			m.activeView = ViewAuth
			return m, nil
		case "ctrl+h":
			log.Debug("Help requested", "active_view", m.activeView)
			// Disable/toggle modal if one already active
			if m.activeModal != ModalNone {
				m.activeModal = ModalNone
			} else {
				m.activeModal = ModalHelp
			}
			return m, nil

		// Handle closing modal when esc is pressed if any is active
		case "esc":
			if m.activeModal != ModalNone {
				m.activeModal = ModalNone
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		log.Debug("Window size changed", "width", m.width, "height", m.height)
		m.width = msg.Width
		m.height = msg.Height

		// Propagate new window size to all views so they are aware and can render correctly
		m.authModel.Resize(msg.Width, msg.Height)
		m.helpModel.Resize(msg.Width, msg.Height)
	}

	// Delegate message processing to the active view
	switch m.activeView {
	case ViewAuth:
		return m.updateAuthView(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	// If there is an active modal it takes presedence
	switch m.activeModal {
	case ModalHelp:
		return m.helpModel.View(m.activeView)
	}

	// Else display the actual view
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
	switch typedMsg := msg.(type) {
	case AuthCompletedMsg:
		log.Info("Authentication successful")
		m.config.Auth.Token = typedMsg.Token
		err := config.UpdateConfig(func(conf *config.Config) {
			conf.Auth.Token = typedMsg.Token
		})
		if err != nil {
			log.Warn("Error saving auth token to config.  Will need to reauthenticate when Hisame opens next", "error", err)
		}
		m.authModel.Reset()
		m.activeView = ViewAnimeList
		// TODO: Initialise/load data from AniList
		return m, nil
	case AuthFailedMsg:
		log.Error("Authentication failed", "error", typedMsg.Error)
		m.authModel.Reset()
		// TODO:  Add better error handling when auth fails
		os.Exit(1)
	}

	// Delegate other messages to the model
	authModel, cmd := m.authModel.Update(msg)
	m.authModel = authModel.(*AuthModel)

	return m, cmd
}
