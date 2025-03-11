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
