package models

import (
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
		log.Debug("Window size changed", "old_width", m.width, "new_width", msg.Width, "old_height", m.height, "new_height", msg.Height)
		m.width = msg.Width
		m.height = msg.Height

		// Propagate new window size to all views so they are aware and can render correctly
		m.authModel.Resize(msg.Width, msg.Height)
		m.helpModel.Resize(msg.Width, msg.Height)
		m.animeListModel.Resize(msg.Width, msg.Height)
		m.episodeSelectModel.Resize(msg.Width, msg.Height)

	case EpisodeLoadedMsg:
		if len(msg.Episodes) == 0 {
			log.Warn("No episodes found for anime", "title", msg.Title)
			return m, nil
		}

		log.Info("Episodes loaded", "count", len(msg.Episodes), "title", msg.Title)

		// Set the episodes in the model
		m.episodeSelectModel.SetEpisodes(msg.Episodes, msg.Title)

		// Activate the episode selection modal
		m.activeModal = ModalEpisodeSelect

		m.animeListModel.DisableLoading()

		return m, nil

	case EpisodeLoadErrorMsg:
		log.Error("Failed to load episodes", "error", msg.Error)
		// Could display an error notification here
		return m, nil

	case EpisodeSelectMsg:
		// An episode was selected from the modal
		if msg.Episode != nil {
			log.Info("Episode selected to play",
				"overall_epNum", msg.Episode.OverallEpisodeNumber,
				"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
				"allanime_id", msg.Episode.AllAnimeID,
				"title", msg.Episode.Title)

			// Close the modal
			m.activeModal = ModalNone

			// Delegate to anime list model to handle playing
			return m.updateAnimeListView(msg)
		}
		return m, nil

	case NextEpisodeFoundMsg:
		log.Info("Next episode found in app model",
			"title", msg.Episode.Title,
			"overall_epNum", msg.Episode.OverallEpisodeNumber,
			"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
			"allanime_id", msg.Episode.AllAnimeID,
		)
		// Close any active modal
		m.activeModal = ModalNone
		m.animeListModel.DisableLoading()

		// Delegate to anime list model to handle loading sources
		return m.updateAnimeListView(msg)

	case EpisodeSourcesLoadedMsg:
		// Delegate to anime list model
		return m.updateAnimeListView(msg)

	case PlaybackStartedMsg:
		log.Info("Playback started",
			"title", msg.EpisodeInfo.Title,
			"episode", msg.EpisodeInfo.AllAnimeEpisodeNumber)

		// Close any active modal
		m.activeModal = ModalNone

		// Disable loading state in anime list
		m.animeListModel.DisableLoading()

		return m, nil

	case PlaybackEndedMsg:
		log.Info("Playback ended",
			"title", msg.EpisodeInfo.Title,
			"episode", msg.EpisodeInfo.AllAnimeEpisodeNumber,
			"progress", msg.Progress)

		// Disable loading state in anime list if it's still active
		m.animeListModel.DisableLoading()

		return m, nil

	case PlaybackErrorMsg:
		log.Error("Playback error",
			"title", msg.EpisodeInfo.Title,
			"episode", msg.EpisodeInfo.AllAnimeEpisodeNumber,
			"error", msg.Error)

		// Disable loading state in anime list
		m.animeListModel.DisableLoading()

		return m, nil

	case PlaybackProgressMsg:
		// This is used for future progress tracking feature
		// For now, we can log it at debug level but don't need UI updates
		log.Debug("Playback progress",
			"title", msg.EpisodeInfo.Title,
			"episode", msg.EpisodeInfo.AllAnimeEpisodeNumber,
			"progress", msg.Progress)

		return m, nil

	}

	// Prioritise delegating messages to a modal if one is active
	switch m.activeModal {
	case ModalEpisodeSelect:
		return m.updateEpisodeSelectModal(msg)
	}

	// Delegate message processing to the active view
	switch m.activeView {
	case ViewAuth:
		return m.updateAuthView(msg)
	case ViewAnimeList:
		return m.updateAnimeListView(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	// If there is an active modal it takes presedence
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

		// Initialize AniList client and services
		client, err := anilist.NewClient(typedMsg.Token)
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
	case AuthFailedMsg:
		log.Error("Authentication failed", "error", typedMsg.Error)
		m.authModel.Reset()
		// TODO:  Add better error handling when auth fails
		return m, tea.Quit
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
