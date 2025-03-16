package domain

// MediaStatus represents which list the anime is in
type MediaStatus string

const (
	StatusCurrent   MediaStatus = "CURRENT"
	StatusPlanning  MediaStatus = "PLANNING"
	StatusCompleted MediaStatus = "COMPLETED"
	StatusDropped   MediaStatus = "DROPPED"
	StatusPaused    MediaStatus = "PAUSED"
	StatusRepeating MediaStatus = "REPEATING"
)

// Anime represents the core anime information
type Anime struct {
	ID           int
	Title        AnimeTitle
	CoverImage   string
	Episodes     int
	NextAiringEp *AiringSchedule
	Status       string
	Format       string
	Season       string
	SeasonYear   string
	AverageScore float64
	UserData     *UserAnimeData
}

// AnimeTitle contains various versions of the anime title
type AnimeTitle struct {
	Romaji  string
	English string
	Native  string
}

// AiringSchedule represents information about an upcoming episode
type AiringSchedule struct {
	Episode      int
	AiringAt     int64
	TimeUntilAir int64
}

// UserAnimeData represents user-specific data for an anime
type UserAnimeData struct {
	Status    MediaStatus
	Score     float64
	Progress  int
	StartDate string
	EndDate   string
	Notes     string
}

// Preferred returns the anime title in the user's preferred language.
// It follows a fallback order if the preferred title format is unavailable:
//   - For "english" preference: English → Romaji → Native
//   - For "romaji" preference: Romaji → English → Native
//   - For "native" preference: Native → Romaji → English
//
// It will return an empty string only if all title formats are empty.
func (at AnimeTitle) Preferred(preference string) string {
	switch preference {
	case "romaji":
		return getFirstNonEmpty(at.Romaji, at.English, at.Native)
	case "english":
		return getFirstNonEmpty(at.English, at.Romaji, at.Native)
	case "native":
		return getFirstNonEmpty(at.Native, at.Romaji, at.English)
	default: // Default to English preference if unspecified
		return getFirstNonEmpty(at.English, at.Romaji, at.Native)
	}
}

// getFirstNonEmpty returns the first non-empty string from the provided arguments
// or an empty string if all arguments are empty
func getFirstNonEmpty(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}
