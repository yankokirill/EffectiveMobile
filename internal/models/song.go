package models

type SongInfo struct {
	ID          int    `json:"id"`
	Title       string `json:"song"`
	Group       string `json:"group"`
	ReleaseDate string `json:"releaseDate"`
	Link        string `json:"link"`
}

type Song struct {
	SongInfo
	Lyrics string
}
