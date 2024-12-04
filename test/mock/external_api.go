package mock

import (
	_ "embed"
	"encoding/json"
	"github.com/yankokirill/song-library/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
)

type SongInfo struct {
	Title string `json:"title"`
	Group string `json:"group"`
}

type Song struct {
	SongInfo
	models.SongDetail
	Valid bool `json:"valid"`
}

//go:embed songs.json
var songsData string
var globalSongStore = make(map[SongInfo]Song)

func LoadSongs() error {
	var songs []Song

	err := json.Unmarshal([]byte(songsData), &songs)
	if err != nil {
		log.Fatalf("Error unmarshaling songsData: %v", err)
	}

	for _, song := range songs {
		songTitle := SongInfo{
			Title: song.Title,
			Group: song.Group,
		}
		globalSongStore[songTitle] = song
	}
	return nil
}

func NewExternalApiServer() *httptest.Server {
	if err := LoadSongs(); err != nil {
		log.Fatalf("failed to unmarshal songs.json: %v", err)
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		songInfo := SongInfo{
			Title: r.URL.Query().Get("song"),
			Group: r.URL.Query().Get("group"),
		}
		if songInfo.Title == "" || songInfo.Group == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		song, ok := globalSongStore[songInfo]
		if !ok {
			http.Error(w, "Song Not Found", http.StatusNotFound)
			return
		}

		if !song.Valid {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(song.SongDetail)
	})

	return httptest.NewServer(handler)
}
