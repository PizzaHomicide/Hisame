package models

// anime_list.go contains the core structure and functionality of the AnimeListModel.
// It defines the main model type, initialization, and essential methods that aren't
// specific to rendering, filtering, input handling, or playback.

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/domain"
	"github.com/PizzaHomicide/hisame/internal/player"
	"github.com/PizzaHomicide/hisame/internal/service"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/components"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

// AnimeFilterSet represents a collection of filters to apply to the anime list
type AnimeFilterSet struct {
	statusFilters        []domain.MediaStatus // Empty slice means no status filter
	hasAvailableEpisodes bool                 // Filter to only anime with aired but unwatched episodes
	isFinishedAiring     bool                 // Filter to anime that have fully completed airing
	searchQuery          string               // Fuzzy search query to match titles against
}

// AnimeListModel handles displaying and interacting with the anime list
type AnimeListModel struct {
	config               *config.Config
	animeService         *service.AnimeService
	playerService        *player.PlayerService
	width, height        int
	loading              bool
	loadingMsg           string
	loadError            error
	spinner              spinner.Model
	filters              AnimeFilterSet
	cursor               int
	allAnime             []*domain.Anime // All anime from the service
	filteredAnime        []*domain.Anime // Anime after applying filters
	searchInput          textinput.Model
	searchMode           bool // Whether we're in search input mode
	playbackCompletionCh chan PlaybackCompletedMsg
}

// NewAnimeListModel creates a new anime list model
func NewAnimeListModel(cfg *config.Config, animeService *service.AnimeService) *AnimeListModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	// Default filters - initially show only CURRENT anime
	defaultFilters := AnimeFilterSet{
		statusFilters: []domain.MediaStatus{domain.StatusCurrent},
	}

	ti := textinput.New()
	ti.Placeholder = "Search anime..."
	ti.Width = 30

	return &AnimeListModel{
		config:               cfg,
		animeService:         animeService,
		playerService:        player.NewPlayerService(cfg),
		loading:              false,
		spinner:              s,
		filters:              defaultFilters,
		cursor:               0,
		allAnime:             []*domain.Anime{},
		filteredAnime:        []*domain.Anime{},
		searchInput:          ti,
		searchMode:           false,
		playbackCompletionCh: make(chan PlaybackCompletedMsg),
	}
}

func (m *AnimeListModel) ViewType() View {
	return ViewAnimeList
}

// Resize updates the model with new dimensions
func (m *AnimeListModel) Resize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the model
func (m *AnimeListModel) Init() tea.Cmd {
	return func() tea.Msg {
		return LoadingMsg{
			Type:        LoadingStart,
			Message:     "Loading anime list...",
			Title:       "Starting Hisame",
			ContextInfo: "Fetching your anime data from AniList",
			Operation:   m.fetchAnimeListCmd(),
		}
	}
}

// The fetchAnimeListCmd creates a command to run in the background
func (m *AnimeListModel) fetchAnimeListCmd() tea.Cmd {
	return func() tea.Msg {
		// Fetch data from service
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := m.animeService.LoadAnimeList(ctx); err != nil {
			return AnimeListLoadResultMsg{
				Success: false,
				Error:   err,
			}
		}

		return AnimeListLoadResultMsg{
			Success:   true,
			AnimeList: m.animeService.GetAnimeList(),
		}
	}
}

func (m *AnimeListModel) HandleAnimeListLoaded(animeList []*domain.Anime) (Model, tea.Cmd) {
	m.allAnime = animeList
	m.applyFilters()
	return m, nil
}

func (m *AnimeListModel) HandleAnimeListError(err error) (Model, tea.Cmd) {
	// TODO:  UX for error here?
	return m, nil
}

// View renders the anime list model
func (m *AnimeListModel) View() string {
	if m.loading {
		return styles.CenteredView(
			m.width,
			m.height,
			fmt.Sprintf("%s %s", m.spinner.View(), m.loadingMsg),
		)
	}

	if m.loadError != nil {
		errorMsg := fmt.Sprintf("Error loading anime list: %v\n\nPress 'r' to retry.", m.loadError)
		return styles.CenteredView(
			m.width,
			m.height,
			styles.ContentBox(m.width-20, errorMsg, 1),
		)
	}

	// Define keybindings to be displayed in footer
	keyBindings := []components.KeyBinding{
		{"↑/↓", "Navigate"},
		{"Enter", "Play next ep"},
		{"Ctrl+p", "Select ep"},
		{"+/-", "Adjust progress"},
		{"d", "Details"},
		{"Ctrl+h", "Help"},
		{"Ctrl+c", "Quit"},
	}

	// Build the view
	header := styles.Header(m.width, "Hisame - Anime List")
	filterStatus := m.renderFilterStatus()
	content := m.renderAnimeList()
	keyBar := components.KeyBindingsBar(m.width, keyBindings)

	if m.searchMode {
		// Show search input at the top of the content
		searchPrompt := styles.Title.Render("Search: ") + m.searchInput.View()
		content = lipgloss.JoinVertical(lipgloss.Left, searchPrompt, content)
	}

	// Layout the components
	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		header,
		filterStatus,
		content,
		styles.CenteredText(m.width, keyBar))
}

// getSelectedAnime returns the currently selected anime or nil if none
func (m *AnimeListModel) getSelectedAnime() *domain.Anime {
	animeList := m.filteredAnime
	if len(animeList) == 0 || m.cursor >= len(animeList) {
		return nil
	}
	return animeList[m.cursor]
}

// DisableLoading disables the loading state
func (m *AnimeListModel) DisableLoading() {
	m.loading = false
}
