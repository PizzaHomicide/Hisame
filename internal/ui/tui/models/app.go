package models

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/repository/anilist"
	"github.com/PizzaHomicide/hisame/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the main application model that coordinates all child models.  It is the high level wrapper.
type AppModel struct {
	config        *config.Config
	activeView    View  // Track the current active 'main view'
	activeModal   Modal // Track the current active 'modal overlay' if any
	width, height int

	// Models used for various views
	authModel          *AuthModel
	animeListModel     *AnimeListModel
	helpModel          *HelpModel
	episodeSelectModel *EpisodeSelectModel

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
		config:             cfg,
		activeView:         initialView,
		activeModal:        ModalNone,
		authModel:          NewAuthModel(),
		animeListModel:     NewAnimeListModel(cfg, animeService),
		helpModel:          NewHelpModel(),
		episodeSelectModel: NewEpisodeSelectModel(),
		animeService:       animeService,
	}
}

func (m AppModel) Init() tea.Cmd {
	log.Info("Initialising Hisame TUI")

	// If starting application on anime list view, load the anime now
	if m.activeView == ViewAnimeList {
		log.Debug("Existing auth detected.  Loading anime list immediately")
		return m.animeListModel.Init()
	}

	return nil
}

func (m AppModel) logMsg(msg tea.Msg) {
	// Log the message type for tracing
	msgType := fmt.Sprintf("%T", msg)
	msgValue := fmt.Sprintf("%+v", msg)

	if msgType == "spinner.TickMsg" {
		// These are too spammy even for trace logging
		return
	}

	// Truncate the value if it's too long
	if len(msgValue) > 100 {
		msgValue = msgValue[:100] + "..."
	}

	log.Trace("Received message in AppModel.Update",
		"type", msgType,
		"value", msgValue,
		"active_modal", m.activeModal,
		"active_view", m.activeView)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logMsg(msg)
	// Handle common message types first
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	}

	// Handle orchestration messages (those that affect multiple components)
	if updatedModel, cmd := m.handleOrchestrationMsg(msg); (updatedModel != AppModel{}) {
		return updatedModel, cmd
	}

	// Any other types of messages should be propagated
	// Prioritize modal handling
	if m.activeModal != ModalNone {
		return m.handleModalMsg(msg)
	}

	// Handle view-specific messages
	return m.handleViewMsg(msg)
}

func (m AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	log.Trace("Handling key message", "key", msg.String())
	// Global key shortcuts that apply everywhere
	switch msg.String() {
	case "ctrl+c":
		log.Info("Quit command received. Shutting down...")
		return m, tea.Quit
	case "ctrl+l":
		return m.handleLogout()
	case "ctrl+h":
		return m.toggleHelpModal()
	case "esc":
		if m.activeModal != ModalNone {
			m.activeModal = ModalNone
			return m, nil
		}
	}

	// Any non-global key messages should be delegated to the active view
	// IPrioritise
	if m.activeModal != ModalNone {
		log.Trace("Delegating key press to modal", "key", msg.String())
		return m.handleModalMsg(msg)
	}

	// Delegate to the active view
	log.Trace("Delegating key press to view", "key", msg.String())
	return m.handleViewMsg(msg)
}

func (m AppModel) handleModalMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeModal {
	case ModalEpisodeSelect:
		return m.updateEpisodeSelectModal(msg)
	case ModalHelp:
		// If needed, handle Help modal messages
		return m, nil
	}
	return m, nil
}

func (m AppModel) handleViewMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeView {
	case ViewAuth:
		return m.updateAuthView(msg)
	case ViewAnimeList:
		return m.updateAnimeListView(msg)
	}
	return m, nil
}

func (m AppModel) handleOrchestrationMsg(msg tea.Msg) (AppModel, tea.Cmd) {
	switch msg := msg.(type) {
	case EpisodeMsg:
		// Handle only episode messages that require orchestration
		switch msg.Type {
		case EpisodeEventLoaded:
			if len(msg.Episodes) == 0 {
				log.Warn("No episodes found for anime", "title", msg.Title)
				m.animeListModel.DisableLoading()
				return m, nil
			}

			log.Info("Episodes loaded", "count", len(msg.Episodes), "title", msg.Title)
			m.episodeSelectModel.SetEpisodes(msg.Episodes, msg.Title)
			m.activeModal = ModalEpisodeSelect
			log.Debug(string("Current active modal: " + m.activeModal))
			m.animeListModel.DisableLoading()
			return m, nil
		case EpisodeEventSelected:
			if msg.Episode != nil {
				log.Info("Episode selected from modal",
					"overall_epNum", msg.Episode.OverallEpisodeNumber,
					"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
					"allanime_id", msg.Episode.AllAnimeID,
					"title", msg.Episode.Title)

				// Close the modal first
				m.activeModal = ModalNone

				// Start loading the sources by delegating to the anime list model
				animeListModel, cmd := m.animeListModel.Update(msg)
				m.animeListModel = animeListModel.(*AnimeListModel)

				return m, cmd
			}
			return m, nil
		}

	case PlaybackMsg:
		// Handle only playback messages that require orchestration
		switch msg.Type {
		case PlaybackEventStarted, PlaybackEventEnded, PlaybackEventError:
			// Close any active modal
			m.activeModal = ModalNone
			m.animeListModel.DisableLoading()
			return m, nil
		}
	}

	// No orchestration was handled - return a zero value to indicate that
	return AppModel{}, nil
}

func (m AppModel) handleLogout() (tea.Model, tea.Cmd) {
	log.Info("Logging out. Cleaning up token from config file...")
	m.config.Auth.Token = ""
	err := config.UpdateConfig(func(conf *config.Config) {
		conf.Auth.Token = ""
	})
	if err != nil {
		log.Warn("Error cleaning up token from config file. May need to manually edit config to remove the token", "error", err)
	}
	// Throw back to login screen
	m.authModel.Reset()
	m.activeView = ViewAuth
	return m, nil
}

func (m AppModel) toggleHelpModal() (tea.Model, tea.Cmd) {
	log.Debug("Help requested", "active_view", m.activeView)
	if m.activeModal != ModalNone {
		m.activeModal = ModalNone
	} else {
		m.activeModal = ModalHelp
	}
	return m, nil
}

func (m AppModel) handleAuthMsg(msg AuthMsg) (tea.Model, tea.Cmd) {
	if msg.Success {
		// Handle successful auth
		return m.handleSuccessfulAuth(msg.Token)
	} else {
		// Handle auth failure
		log.Error("Authentication failed", "error", msg.Error)
		m.authModel.Reset()
		return m, tea.Quit
	}
}

// handleWindowSizeMsg updates all models with the new window dimensions
func (m AppModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	log.Debug("Window size changed", "old_width", m.width, "new_width", msg.Width,
		"old_height", m.height, "new_height", msg.Height)

	// Update the app model's dimensions
	m.width = msg.Width
	m.height = msg.Height

	// Propagate new window size to all views so they can render correctly
	m.authModel.Resize(msg.Width, msg.Height)
	m.helpModel.Resize(msg.Width, msg.Height)
	m.animeListModel.Resize(msg.Width, msg.Height)
	m.episodeSelectModel.Resize(msg.Width, msg.Height)

	return m, nil
}

// handleSuccessfulAuth processes a successful authentication
func (m AppModel) handleSuccessfulAuth(token string) (tea.Model, tea.Cmd) {
	log.Info("Authentication successful")

	// Save the token to the config
	m.config.Auth.Token = token
	err := config.UpdateConfig(func(conf *config.Config) {
		conf.Auth.Token = token
	})
	if err != nil {
		log.Warn("Error saving auth token to config. Will need to reauthenticate when Hisame opens next", "error", err)
	}

	// Reset the auth model for future use
	m.authModel.Reset()

	// Initialize AniList client and services
	client, err := anilist.NewClient(token)
	if err != nil {
		log.Error("Failed to create AniList client after authentication", "error", err)
		return m, tea.Quit
	}

	// Set up the anime service and models
	animeRepo := anilist.NewAnimeRepository(client)
	m.animeService = service.NewAnimeService(animeRepo)
	m.animeListModel = NewAnimeListModel(m.config, m.animeService)
	m.animeListModel.Resize(m.width, m.height)

	// Change the active view
	m.activeView = ViewAnimeList

	// Initialize the anime list view
	return m, m.animeListModel.Init()
}

func (m AppModel) View() string {
	// If there is an active modal it takes precedence
	switch m.activeModal {
	case ModalHelp:
		return m.helpModel.View(m.activeView)
	case ModalEpisodeSelect:
		return m.episodeSelectModel.View()
	}

	// Else display the actual view
	switch m.activeView {
	case ViewAuth:
		return m.authModel.View()
	case ViewAnimeList:
		return m.animeListModel.View()
	default:
		log.Warn("Unknown view", "view", m.activeView)
		return "Unknown view\nPress ctrl+c to quit."
	}
}

// updateAuthView delegates message processing to
func (m AppModel) updateAuthView(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Any messages that require orchestration/view changing specific to the auth view
	switch msg := msg.(type) {
	case AuthMsg:
		if msg.Success {
			log.Info("Authentication successful")
			m.config.Auth.Token = msg.Token
			err := config.UpdateConfig(func(conf *config.Config) {
				conf.Auth.Token = msg.Token
			})
			if err != nil {
				log.Warn("Error saving auth token to config. Will need to reauthenticate when Hisame opens next", "error", err)
			}
			m.authModel.Reset()

			// Initialize AniList client and services
			client, err := anilist.NewClient(msg.Token)
			if err != nil {
				log.Error("Failed to create AniList client after authentication", "error", err)
				return m, tea.Quit
			}

			animeRepo := anilist.NewAnimeRepository(client)
			m.animeService = service.NewAnimeService(animeRepo)
			m.animeListModel = NewAnimeListModel(m.config, m.animeService)
			m.animeListModel.Resize(m.width, m.height)
			m.activeView = ViewAnimeList

			return m, m.animeListModel.Init()
		} else {
			log.Error("Authentication failed", "error", msg.Error)
			m.authModel.Reset()
			// TODO: Add better error handling when auth fails
			return m, tea.Quit
		}
	}

	// Delegate other messages to the model
	authModel, cmd := m.authModel.Update(msg)
	m.authModel = authModel.(*AuthModel)

	return m, cmd
}

// updateAnimeListView delegates message processing to the anime list model
func (m AppModel) updateAnimeListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to the animeListModel
	animeListModel, cmd := m.animeListModel.Update(msg)
	m.animeListModel = animeListModel.(*AnimeListModel)

	return m, cmd
}

func (m AppModel) updateEpisodeSelectModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to the episode select model
	model, cmd := m.episodeSelectModel.Update(msg)
	m.episodeSelectModel = model.(*EpisodeSelectModel)

	return m, cmd
}
