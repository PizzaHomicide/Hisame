package keybindings

// Action represents a specific action that can be triggered by a key
type Action string

// Define all possible actions
const (
	// Global actions
	ActionQuit       Action = "quit"
	ActionToggleHelp Action = "toggle_help"
	ActionLogout     Action = "logout"
	ActionBack       Action = "back" // General purpose "go back" or "cancel"

	// Auth view actions
	ActionLogin Action = "login"

	// Anime list actions
	ActionMoveUp                Action = "move_up"
	ActionMoveDown              Action = "move_down"
	ActionPageUp                Action = "page_up"
	ActionPageDown              Action = "page_down"
	ActionViewAnimeDetail       Action = "view_anime_detail"
	ActionRefreshAnimeList      Action = "refresh_anime_list"
	ActionPlayNextEpisode       Action = "play_next_episode"
	ActionChooseEpisode         Action = "choose_episode"
	ActionToggleSearch          Action = "toggle_search"
	ActionIncrementProgress     Action = "increment_progress"
	ActionDecrementProgress     Action = "decrement_progress"
	ActionToggleStatusCurrent   Action = "toggle_status_current"
	ActionToggleStatusPlanning  Action = "toggle_status_planning"
	ActionToggleStatusComplete  Action = "toggle_status_complete"
	ActionToggleStatusDropped   Action = "toggle_status_dropped"
	ActionToggleStatusPaused    Action = "toggle_status_paused"
	ActionToggleStatusRepeating Action = "toggle_status_repeating"
	ActionToggleNewEpisodes     Action = "toggle_new_episodes"
	ActionToggleFinishedAiring  Action = "toggle_finished_airing"
)

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

// GlobalBindings contains key bindings that work across all views
var GlobalBindings = []Binding{
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

// AuthBindings contains key bindings specific to the auth view
var AuthBindings = []Binding{
	{
		Action: ActionLogin,
		KeyMap: KeyMap{
			Primary:   "enter",
			Secondary: "l",
			Help:      "Start login process",
		},
	},
}

// AnimeListBindings contains key bindings specific to the anime list view
var AnimeListBindings = []Binding{
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
		Action: ActionViewAnimeDetail,
		KeyMap: KeyMap{
			Primary: "enter",
			Help:    "View anime details",
		},
	},
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
			Primary: "p",
			Help:    "Play next episode",
		},
	},
	{
		Action: ActionChooseEpisode,
		KeyMap: KeyMap{
			Primary: "ctrl+p",
			Help:    "Choose episode to play",
		},
	},
	{
		Action: ActionToggleSearch,
		KeyMap: KeyMap{
			Primary: "ctrl+f",
			Help:    "Search anime",
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
		Action: ActionToggleStatusCurrent,
		KeyMap: KeyMap{
			Primary: "1",
			Help:    "Toggle watching filter",
		},
	},
	{
		Action: ActionToggleStatusPlanning,
		KeyMap: KeyMap{
			Primary: "2",
			Help:    "Toggle planning filter",
		},
	},
	{
		Action: ActionToggleStatusComplete,
		KeyMap: KeyMap{
			Primary: "3",
			Help:    "Toggle completed filter",
		},
	},
	{
		Action: ActionToggleStatusDropped,
		KeyMap: KeyMap{
			Primary: "4",
			Help:    "Toggle dropped filter",
		},
	},
	{
		Action: ActionToggleStatusPaused,
		KeyMap: KeyMap{
			Primary: "5",
			Help:    "Toggle on-hold filter",
		},
	},
	{
		Action: ActionToggleStatusRepeating,
		KeyMap: KeyMap{
			Primary: "6",
			Help:    "Toggle repeating filter",
		},
	},
	{
		Action: ActionToggleNewEpisodes,
		KeyMap: KeyMap{
			Primary: "a",
			Help:    "Toggle available episodes filter",
		},
	},
	{
		Action: ActionToggleFinishedAiring,
		KeyMap: KeyMap{
			Primary: "f",
			Help:    "Toggle finished airing filter",
		},
	},
}

// EpisodeSelectBindings contains key bindings specific to the episode selection view
var EpisodeSelectBindings = []Binding{
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
		Action: ActionViewAnimeDetail,
		KeyMap: KeyMap{
			Primary: "enter",
			Help:    "Select episode",
		},
	},
	{
		Action: ActionToggleSearch,
		KeyMap: KeyMap{
			Primary: "ctrl+f",
			Help:    "Search episodes",
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
func GetActionByKey(key string, bindings []Binding) Action {
	for _, binding := range bindings {
		if binding.KeyMap.Primary == key || binding.KeyMap.Secondary == key {
			return binding.Action
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
