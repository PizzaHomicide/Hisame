package keybindings

import tea "github.com/charmbracelet/bubbletea"

// Action represents a specific action that can be triggered by a key
type Action string

// Define all possible actions
const (
	// Global actions
	ActionQuit       Action = "quit"
	ActionToggleHelp Action = "toggle_help"
	ActionLogout     Action = "logout"
	ActionBack       Action = "back" // General purpose "go back" or "cancel"

	// Navigation actions
	ActionMoveUp     Action = "move_up"
	ActionMoveDown   Action = "move_down"
	ActionPageUp     Action = "page_up"
	ActionPageDown   Action = "page_down"
	ActionMoveTop    Action = "move_top"
	ActionMoveBottom Action = "move_bottom"

	// Auth view actions
	ActionLogin Action = "login"

	// Anime list actions
	ActionSelectEpisode               Action = "select_episode"
	ActionRefreshAnimeList            Action = "refresh_anime_list"
	ActionPlayNextEpisode             Action = "play_next_episode"
	ActionOpenEpisodeSelector         Action = "episode_selector"
	ActionIncrementProgress           Action = "increment_progress"
	ActionDecrementProgress           Action = "decrement_progress"
	ActionToggleFilterStatusCurrent   Action = "toggle_filter_status_current"
	ActionToggleFilterStatusPlanning  Action = "toggle_filter_status_planning"
	ActionToggleFilterStatusComplete  Action = "toggle_filter_status_complete"
	ActionToggleFilterStatusDropped   Action = "toggle_filter_status_dropped"
	ActionToggleFilterStatusPaused    Action = "toggle_filter_status_paused"
	ActionToggleFilterStatusRepeating Action = "toggle_filter_status_repeating"
	ActionToggleFilterNewEpisodes     Action = "toggle_filter_new_episodes"
	ActionToggleFilterFinishedAiring  Action = "toggle_filter_finished_airing"

	// Search mode actions
	ActionEnableSearch   Action = "enable_search"
	ActionSearchComplete Action = "search_complete"
)

// ContextName represents a specific UI context in the application that has its own keybinds
type ContextName string

const (
	ContextGlobal           ContextName = "global"
	ContextAuth             ContextName = "auth"
	ContextAnimeList        ContextName = "anime_list"
	ContextEpisodeSelection ContextName = "episode_selection"
	ContextSearchMode       ContextName = "search_mode"
	ContextHelp             ContextName = "help"
)

var ContextBindings = map[ContextName][]Binding{
	ContextGlobal:           globalBindings,
	ContextAuth:             authBindings,
	ContextAnimeList:        animeListBindings,
	ContextEpisodeSelection: episodeSelectBindings,
	ContextSearchMode:       searchModeBindings,
	ContextHelp:             helpBindings,
}

// KeyMap stores the mappings from actions to key sequences for each context
type KeyMap struct {
	Primary   string
	Secondary string // Optional alternative key
	Help      string // Description for help screen
}

// Binding maps an action to its keys and help text
type Binding struct {
	Action Action
	KeyMap KeyMap
}

// navigationBindings contains general navigation bindings for consistent navigation across the app
var navigationBindings = []Binding{
	{
		Action: ActionMoveUp,
		KeyMap: KeyMap{
			Primary:   "up",
			Secondary: "k",
			Help:      "Move cursor up",
		},
	},
	{
		Action: ActionMoveDown,
		KeyMap: KeyMap{
			Primary:   "down",
			Secondary: "j",
			Help:      "Move cursor down",
		},
	},
	{
		Action: ActionPageUp,
		KeyMap: KeyMap{
			Primary: "pgup",
			Help:    "Move up one page",
		},
	},
	{
		Action: ActionPageDown,
		KeyMap: KeyMap{
			Primary: "pgdown",
			Help:    "Move down one page",
		},
	},
	{
		Action: ActionMoveTop,
		KeyMap: KeyMap{
			Primary: "home",
			Help:    "Move top of view",
		},
	},
	{
		Action: ActionMoveBottom,
		KeyMap: KeyMap{
			Primary: "end",
			Help:    "Move bottom of view",
		},
	},
}

// globalBindings contains key bindings that work across all views
var globalBindings = []Binding{
	{
		Action: ActionQuit,
		KeyMap: KeyMap{
			Primary: "ctrl+c",
			Help:    "Quit application",
		},
	},
	{
		Action: ActionToggleHelp,
		KeyMap: KeyMap{
			Primary: "ctrl+h",
			Help:    "Toggle help screen",
		},
	},
	{
		Action: ActionLogout,
		KeyMap: KeyMap{
			Primary: "ctrl+l",
			Help:    "Logout (clear token)",
		},
	},
	{
		Action: ActionBack,
		KeyMap: KeyMap{
			Primary: "esc",
			Help:    "Go back/cancel current action",
		},
	},
}

// authBindings contains key bindings specific to the auth view
var authBindings = []Binding{
	{
		Action: ActionLogin,
		KeyMap: KeyMap{
			Primary:   "enter",
			Secondary: "l",
			Help:      "Start login process",
		},
	},
}

// helpBindings contains key bindings specific to the help view
var helpBindings = withNavigation([]Binding{})

// animeListBindings contains key bindings specific to the anime list view
var animeListBindings = withNavigation([]Binding{

	{
		Action: ActionRefreshAnimeList,
		KeyMap: KeyMap{
			Primary: "r",
			Help:    "Refresh anime list",
		},
	},
	{
		Action: ActionPlayNextEpisode,
		KeyMap: KeyMap{
			Primary:   "enter",
			Secondary: "p",
			Help:      "Play next episode",
		},
	},
	{
		Action: ActionOpenEpisodeSelector,
		KeyMap: KeyMap{
			Primary: "ctrl+p",
			Help:    "Choose episode to play",
		},
	},
	{
		Action: ActionEnableSearch,
		KeyMap: KeyMap{
			Primary:   "/",
			Secondary: "ctrl+f",
			Help:      "Search anime",
		},
	},
	{
		Action: ActionIncrementProgress,
		KeyMap: KeyMap{
			Primary: "+",
			Help:    "Increment episode progress",
		},
	},
	{
		Action: ActionDecrementProgress,
		KeyMap: KeyMap{
			Primary: "-",
			Help:    "Decrement episode progress",
		},
	},
	// Filters
	{
		Action: ActionToggleFilterStatusCurrent,
		KeyMap: KeyMap{
			Primary: "1",
			Help:    "Toggle watching filter",
		},
	},
	{
		Action: ActionToggleFilterStatusPlanning,
		KeyMap: KeyMap{
			Primary: "2",
			Help:    "Toggle planning filter",
		},
	},
	{
		Action: ActionToggleFilterStatusComplete,
		KeyMap: KeyMap{
			Primary: "3",
			Help:    "Toggle completed filter",
		},
	},
	{
		Action: ActionToggleFilterStatusDropped,
		KeyMap: KeyMap{
			Primary: "4",
			Help:    "Toggle dropped filter",
		},
	},
	{
		Action: ActionToggleFilterStatusPaused,
		KeyMap: KeyMap{
			Primary: "5",
			Help:    "Toggle on-hold filter",
		},
	},
	{
		Action: ActionToggleFilterStatusRepeating,
		KeyMap: KeyMap{
			Primary: "6",
			Help:    "Toggle repeating filter",
		},
	},
	{
		Action: ActionToggleFilterNewEpisodes,
		KeyMap: KeyMap{
			Primary: "a",
			Help:    "Toggle available episodes filter",
		},
	},
	{
		Action: ActionToggleFilterFinishedAiring,
		KeyMap: KeyMap{
			Primary: "f",
			Help:    "Toggle finished airing filter",
		},
	},
})

// episodeSelectBindings contains key bindings specific to the episode selection view
var episodeSelectBindings = withNavigation([]Binding{
	{
		Action: ActionSelectEpisode,
		KeyMap: KeyMap{
			Primary: "enter",
			Help:    "Select episode",
		},
	},
	{
		Action: ActionEnableSearch,
		KeyMap: KeyMap{
			Primary:   "/",
			Secondary: "ctrl+f",
			Help:      "Search episodes",
		},
	},
})

// searchModeBindings contains key bindings specific for when search mode is active
var searchModeBindings = []Binding{
	{
		Action: ActionBack,
		KeyMap: KeyMap{
			Primary:   "esc",
			Secondary: "ctrl+f",
			Help:      "Exit search mode and remove the filter",
		},
	},
	{
		Action: ActionSearchComplete,
		KeyMap: KeyMap{
			Primary: "enter",
			Help:    "Apply the search filter and return control to the original view",
		},
	},
}

// GetActionKey returns the primary key for an action
func GetActionKey(action Action, bindings []Binding) string {
	for _, binding := range bindings {
		if binding.Action == action {
			return binding.KeyMap.Primary
		}
	}
	return ""
}

// GetActionSecondaryKey returns the secondary key for an action if it exists
func GetActionSecondaryKey(action Action, bindings []Binding) string {
	for _, binding := range bindings {
		if binding.Action == action {
			return binding.KeyMap.Secondary
		}
	}
	return ""
}

// GetBindingByKey returns the action and help text for a given key
func GetBindingByKey(key string, bindings []Binding) (Action, string) {
	for _, binding := range bindings {
		if binding.KeyMap.Primary == key || binding.KeyMap.Secondary == key {
			return binding.Action, binding.KeyMap.Help
		}
	}
	return "", ""
}

// GetActionByKey returns just the action for a given key, or an empty Action if not found
func GetActionByKey(keyMsg tea.KeyMsg, name ContextName) Action {
	if bindings, exists := ContextBindings[name]; exists {
		key := keyMsg.String()
		for _, binding := range bindings {
			if binding.KeyMap.Primary == key || binding.KeyMap.Secondary == key {
				return binding.Action
			}
		}
	}
	return ""
}

// FormatKeyHelp formats a key binding for display in help text
func FormatKeyHelp(binding Binding) string {
	if binding.KeyMap.Secondary != "" {
		return binding.KeyMap.Primary + "/" + binding.KeyMap.Secondary + ": " + binding.KeyMap.Help
	}
	return binding.KeyMap.Primary + ": " + binding.KeyMap.Help
}

// GetHelpText generates formatted help text for a set of bindings
func GetHelpText(title string, bindings []Binding) string {
	helpText := "## " + title + "\n\n"
	for _, binding := range bindings {
		helpText += "* " + FormatKeyHelp(binding) + "\n"
	}
	return helpText
}

// withNavigation is a helper function to include navigation bindings in other binding sets
func withNavigation(bindings []Binding) []Binding {
	return append(append([]Binding{}, navigationBindings...), bindings...)
}
