package models

type PaginationInfo struct {
	PrevGroup string `json:"prevGroup"`
	PrevSong  string `json:"prevSong"`
	Limit     int    `json:"limit"`
}
