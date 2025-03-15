package domain

type User struct {
	ID         int
	Name       string
	Avatar     string
	SiteURL    string
	Statistics UserStatistics
}

type UserStatistics struct {
	AnimeCount      int
	MangaCount      int
	EpisodesWatched int
	ChaptersRead    int
}
