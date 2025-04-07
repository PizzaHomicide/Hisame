package models

import (
	"errors"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/repository/anilist"
	"github.com/PizzaHomicide/hisame/internal/service"
	kb "github.com/PizzaHomicide/hisame/internal/ui/tui/keybindings"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

// Model is the interface that all our models should implement
type Model interface {
	// Init initializes the model and returns any initial command
	Init() tea.Cmd

	// Update handles messages and returns the updated model and any command
	Update(msg tea.Msg) (Model, tea.Cmd)

	// View renders the model to a string
	View() string

	// Resize updates a models width & height
	Resize(width, height int)

	// ViewType returns the type of the view
	ViewType() View
}

// AppModel is the main application model that coordinates all child models.  It is the high level wrapper.
type AppModel struct {
	config        *config.Config
	modelStack    []Model // UI model stack.  The top model is rendered and handles non-global/orchestration messages
	width, height int

	// Models used for various views
	authModel          *AuthModel
	animeListModel     *AnimeListModel
	helpModel          *HelpModel
	episodeSelectModel *EpisodeSelectModel

	// Services used for fetching and updating state
	animeService *service.AnimeService
}

func NewAppModel(cfg *config.Config) AppModel {
	// Create models
	authModel := NewAuthModel()
	helpModel := NewHelpModel()
	episodeSelectModel := NewEpisodeSelectModel()

	// Create an initial loading model for startup
	initialLoadingModel := NewLoadingModel("Starting Hisame...").
		WithTitle("Initialising")

	// Start with just the loading model
	modelStack := []Model{initialLoadingModel}

	app := AppModel{
		config:             cfg,
		modelStack:         modelStack,
		authModel:          authModel,
		helpModel:          helpModel,
		episodeSelectModel: episodeSelectModel,
	}

	return app
}

// CurrentModel returns the current active model (top of the stack)
func (m AppModel) CurrentModel() Model {
	if len(m.modelStack) == 0 {
		log.Error("Model stack is empty, this should never happen")
		return nil
	}
	return m.modelStack[len(m.modelStack)-1]
}

// PushModel adds a model to the top of the stack and ensures it's properly sized
func (m *AppModel) PushModel(model Model) {
	model.Resize(m.width, m.height)
	// Add to the stack
	m.modelStack = append(m.modelStack, model)
	log.Debug("Pushed model onto stack", "model_type", model.ViewType(), "stack_size", len(m.modelStack))
}

// PopModel removes the top model from the stack
func (m *AppModel) PopModel() {
	if len(m.modelStack) <= 1 {
		log.Warn("Attempted to pop the last model from the stack, ignoring")
		return
	}

	m.modelStack = m.modelStack[:len(m.modelStack)-1]
	log.Debug("Popped model from stack", "new_top", fmt.Sprintf("%T", m.CurrentModel()), "stack_size", len(m.modelStack))
}

// SetStack completely replaces the model stack
func (m *AppModel) SetStack(models []Model) {
	if len(models) == 0 {
		log.Error("Attempted to set empty model stack, ignoring")
		return
	}

	m.modelStack = models

	// Resize all models in the new stack
	for _, model := range m.modelStack {
		if resizable, ok := model.(interface{ Resize(width, height int) }); ok {
			resizable.Resize(m.width, m.height)
		}
	}

	log.Debug("Set new model stack", "top_model", fmt.Sprintf("%T", m.CurrentModel()), "stack_size", len(m.modelStack))
}

func (m AppModel) Init() tea.Cmd {
	log.Info("Initialising Hisame TUI")

	// Start the loading spinner and begin token validation
	return tea.Batch(
		m.CurrentModel().Init(), // Initialize the loading model
		m.validateTokenCmd(),    // Start token validation process
	)
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
		"current_model", m.CurrentModel().ViewType())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logMsg(msg)
	// Handle window size changes globally
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowMsg.Width
		m.height = windowMsg.Height

		// Resize all models in the stack
		for _, model := range m.modelStack {
			model.Resize(m.width, m.height)
		}

		// No need to propagate this message further
		return m, nil
	}

	// Handle global key shortcuts first
	if cmd := m.handleKeyMsg(msg); cmd != nil {
		return m, cmd
	}

	// Handle orchestration messages
	if cmd := m.handleOrchestrationMsg(msg); cmd != nil {
		return m, cmd
	}

	// Update the current model
	currentModel := m.CurrentModel()
	if currentModel == nil {
		log.Error("No current model to update")
		return m, nil
	}

	updatedModel, cmd := currentModel.Update(msg)

	// Replace the current model with the updated one
	if updatedModel != nil {
		m.modelStack[len(m.modelStack)-1] = updatedModel
	}

	return m, cmd
}

func (m *AppModel) handleKeyMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch kb.GetActionByKey(msg.String(), kb.GlobalBindings) {
		case kb.ActionQuit:
			log.Info("Quit command received. Shutting down...")
			return tea.Quit

		case kb.ActionLogout:
			return m.handleLogout()

		case kb.ActionToggleHelp:
			return m.handleToggleHelp()

		case kb.ActionBack:
			// If we have more than one model in the stack, pop the top one
			if len(m.modelStack) > 1 {
				m.PopModel()
				return nil
			}
		}
	}
	return nil
}

// handleOrchestrationMsg handles messages that require coordination between models
func (m *AppModel) handleOrchestrationMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case TokenValidationMsg:
		if !msg.Valid {
			if msg.IsNetwork {
				// Network error - show error and exit
				return func() tea.Msg {
					fmt.Fprintf(os.Stderr, "Network error: %v\nPlease check your connection and try again.\n", msg.Error)
					return tea.Quit()
				}
			}

			// Invalid token - clear it and go to auth screen
			if msg.Error != nil {
				log.Warn("Invalid token in config. Clearing token.", "error", msg.Error)
				m.config.Auth.Token = ""
				err := config.UpdateConfig(func(conf *config.Config) {
					conf.Auth.Token = ""
				})
				if err != nil {
					log.Warn("Failed to clear invalid token from config", "error", err)
				}
			}

			// Go to auth screen
			m.SetStack([]Model{m.authModel})
			return m.authModel.Init()
		}

		// Valid token - set up services and go to anime list
		animeRepo := anilist.NewAnimeRepository(msg.Client)
		animeService := service.NewAnimeService(animeRepo)
		animeListModel := NewAnimeListModel(m.config, animeService)

		// Save references
		m.animeService = animeService
		m.animeListModel = animeListModel

		// Push anime list model
		m.SetStack([]Model{m.animeListModel})

		// Now start loading the anime list data
		return func() tea.Msg {
			return LoadingMsg{
				Type:      LoadingStart,
				Message:   "Loading your anime list...",
				Title:     "Fetching Data",
				Operation: animeListModel.fetchAnimeListCmd(),
			}
		}
	case AuthMsg:
		if msg.Success {
			return m.handleSuccessfulAuth(msg.Token)
		} else {
			log.Error("Authentication failed", "error", msg.Error)
			// Reset auth model in case it's in a bad state
			m.authModel = NewAuthModel()
			m.SetStack([]Model{m.authModel})
			return tea.Quit
		}

	case EpisodeMsg:
		switch msg.Type {
		case EpisodeEventLoaded:
			if len(msg.Episodes) == 0 {
				log.Warn("No episodes found for anime", "title", msg.Title)
				// Turn off loading in anime list model
				m.animeListModel.DisableLoading()
				return nil
			}

			log.Info("Episodes loaded", "count", len(msg.Episodes), "title", msg.Title)
			m.episodeSelectModel.SetEpisodes(msg.Episodes, msg.Title)
			m.PushModel(m.episodeSelectModel)
			m.animeListModel.DisableLoading()
			return nil

		case EpisodeEventSelected:
			if msg.Episode != nil {
				log.Info("Episode selected from episode select model",
					"overall_epNum", msg.Episode.OverallEpisodeNumber,
					"allanime_epNum", msg.Episode.AllAnimeEpisodeNumber,
					"title", msg.Episode.AllAnimeName)

				// Pop episode select model
				m.PopModel()

				// Delegate to anime list model to handle starting playback
				_, cmd := m.animeListModel.Update(msg)
				return cmd
			}

		case EpisodeEventError:
			log.Warn("Could not find episode", "error", msg.Error)
			m.animeListModel.DisableLoading()
			return nil
		}

	case PlaybackMsg:
		// Some playback messages affect the model stack
		switch msg.Type {
		case PlaybackEventStarted, PlaybackEventEnded, PlaybackEventError:
			// Make sure any loading indicators are disabled in the anime list
			m.animeListModel.DisableLoading()
			return nil
		}

	case AnimeListLoadResultMsg:
		if currentModel, ok := m.CurrentModel().(*LoadingModel); ok {
			log.Debug("Stopping loading for anime list refresh",
				"elapsed", currentModel.GetElapsedTime())
			m.PopModel()
		}

		// Then forward the result to the AnimeListModel
		if msg.Success {
			_, cmd := m.animeListModel.HandleAnimeListLoaded(msg.AnimeList)
			return cmd
		} else {
			_, cmd := m.animeListModel.HandleAnimeListError(msg.Error)
			return cmd
		}

	case LoadingMsg:
		switch msg.Type {
		case LoadingStart:
			// Create and push a loading model
			loadingModel := NewLoadingModel(msg.Message)

			// Apply optional configurations if provided
			if msg.Title != "" {
				loadingModel = loadingModel.WithTitle(msg.Title)
			}
			if msg.ContextInfo != "" {
				loadingModel = loadingModel.WithContextInfo(msg.ContextInfo)
			}
			if msg.ActionText != "" {
				loadingModel = loadingModel.WithActionText(msg.ActionText)
			}

			log.Debug("Starting loading state", "message", msg.Message)
			m.PushModel(loadingModel)

			// If there's an operation to run during loading, execute it
			if msg.Operation != nil {
				return tea.Batch(
					loadingModel.Init(),
					msg.Operation,
				)
			}

			return loadingModel.Init()

		case LoadingStop:
			m.popLoadingModel()
			return nil
		}
	}

	return nil
}

func (m *AppModel) popLoadingModel() {
	if currentModel, ok := m.CurrentModel().(*LoadingModel); ok {
		log.Debug("Stopping loading state",
			"message", currentModel.message,
			"elapsed", currentModel.GetElapsedTime())
		m.PopModel()
	} else {
		log.Warn("Received LoadingStop but current model is not a LoadingModel")
	}
}

// handleLogout handles the logout action
func (m *AppModel) handleLogout() tea.Cmd {
	log.Info("Logging out. Cleaning up token from config file...")
	m.config.Auth.Token = ""
	err := config.UpdateConfig(func(conf *config.Config) {
		conf.Auth.Token = ""
	})
	if err != nil {
		log.Warn("Error cleaning up token from config file. May need to manually edit config to remove the token", "error", err)
	}

	// Reset auth model and make it the only model in stack
	m.authModel = NewAuthModel()
	m.SetStack([]Model{m.authModel})

	return nil
}

func (m *AppModel) handleToggleHelp() tea.Cmd {
	// Toggle help screen
	if _, ok := m.CurrentModel().(*HelpModel); ok {
		// Help is already active, pop it
		m.PopModel()
	} else {
		// Get context from current model
		context := ViewAnimeList // Default fallback
		if currentModel := m.CurrentModel(); currentModel != nil {
			context = currentModel.ViewType()
		}

		// Set context and push help model
		m.helpModel.SetContext(context)
		m.PushModel(m.helpModel)
	}
	return nil
}

// handleSuccessfulAuth handles a successful authentication
func (m *AppModel) handleSuccessfulAuth(token string) tea.Cmd {
	log.Info("Authentication successful")

	// Save the token to the config
	m.config.Auth.Token = token
	err := config.UpdateConfig(func(conf *config.Config) {
		conf.Auth.Token = token
	})
	if err != nil {
		log.Warn("Error saving auth token to config. Will need to reauthenticate when Hisame opens next", "error", err)
	}

	// Initialize AniList client and services
	client, err := anilist.NewClient(token)
	if err != nil {
		log.Error("Failed to create AniList client after authentication", "error", err)
		return tea.Quit
	}

	// Set up the anime service and models
	animeRepo := anilist.NewAnimeRepository(client)
	m.animeService = service.NewAnimeService(animeRepo)
	m.animeListModel = NewAnimeListModel(m.config, m.animeService)

	// Replace the entire stack with just the anime list model
	m.SetStack([]Model{m.animeListModel})

	// Initialize the anime list model
	return m.animeListModel.Init()
}

func (m AppModel) View() string {
	// Render the current model
	current := m.CurrentModel()
	if current == nil {
		return "Error: No active model to display\nThis should not happen.  Please exit Hisame with ctrl+c"
	}

	return current.View()
}

func (m AppModel) validateTokenCmd() tea.Cmd {
	return func() tea.Msg {
		token := m.config.Auth.Token

		if token == "" {
			// No token, go straight to auth screen
			return TokenValidationMsg{
				Valid: false,
			}
		}

		// Validate token by making API call
		client, err := anilist.NewClient(token)
		if err != nil {
			// Handle various error types as before
			var netErr anilist.NetworkError
			if errors.As(err, &netErr) {
				return TokenValidationMsg{
					Valid:     false,
					Error:     err,
					IsNetwork: true,
				}
			}

			return TokenValidationMsg{
				Valid: false,
				Error: err,
			}
		}

		// Token is valid
		return TokenValidationMsg{
			Valid:  true,
			Client: client,
		}
	}
}
