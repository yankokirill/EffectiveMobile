package http

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/yankokirill/song-library/internal/models"
	"github.com/yankokirill/song-library/internal/repository/postgres"
	"github.com/yankokirill/song-library/internal/rpc"
	"log"
	"net/http"
	"strconv"
)

type SongLyricsParams struct {
	ID     int
	Offset int
	Limit  int
}

func parseSongLyricsParams(r *http.Request) (*SongLyricsParams, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, fmt.Errorf("'offset' must be a non-negative integer")
		}
	}

	limit := 20
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return nil, fmt.Errorf("'limit' must be a positive integer")
		}
	}

	return &SongLyricsParams{ID: id, Offset: offset, Limit: limit}, nil
}

func parseSongPaginationInfo(r *http.Request) (*models.PaginationInfo, error) {
	groupName := r.URL.Query().Get("prevGroup")
	songTitle := r.URL.Query().Get("prevSong")
	limitStr := r.URL.Query().Get("limit")

	limit := 10
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return nil, fmt.Errorf("invalid 'limit' parameter")
		}
	}

	return &models.PaginationInfo{
		PrevGroup: groupName,
		PrevSong:  songTitle,
		Limit:     limit,
	}, nil
}

func parseID(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return 0, fmt.Errorf("missing 'id' query parameter")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing 'id': %w", err)
	}
	if id <= 0 {
		return 0, fmt.Errorf("'id' query parameter must be a positive integer")
	}
	return id, nil
}

// @Summary Get information about songs
// @Tags API
// @Description Get song information in partitions using lexicographical order and pagination.
// Provide the `prevSong` and `prevGroup` parameters to define the starting point for the next partition.
// If these parameters are not provided, retrieval starts from the first song in the library.
// @Param prevSong query string false "Title of the last song in the previous partition"
// @Param prevGroup query string false "Group of the last song in the previous partition"
// @Param limit query int false "Maximum number of songs to retrieve" default(10)
// @Success 200 {object} []models.SongInfo
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs [get]
func (s *Server) getSongsInfoHandler(w http.ResponseWriter, r *http.Request) {
	hint, err := parseSongPaginationInfo(r)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	songs, err := s.db.GetSongsInfo(hint)
	if err != nil {
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		log.Printf("%v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(songs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("error encoding response: %v", err)
	}
}

// @Summary Get information about songs of a specific group
// @Tags API
// @Description Get song information for a specific musical group in partitions using lexicographical order and pagination.
// Provide the `prevSong` parameter to define the starting point for the next partition.
// If this parameter is not provided, retrieval starts from the first song of the specified group.
// @Param group path string true "Name of the group"
// @Param prevSong query string false "Title of the last song in the previous partition"
// @Param limit query int false "Maximum number of songs to retrieve" default(10)
// @Success 200 {object} []models.SongInfo
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /songs/{group} [get]
func (s *Server) getGroupSongsInfoHandler(w http.ResponseWriter, r *http.Request) {
	group := chi.URLParam(r, "group")
	hint, err := parseSongPaginationInfo(r)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	songs, err := s.db.GetGroupSongsInfo(group, hint)
	if err != nil {
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		log.Printf("%v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(songs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("error encoding response: %v", err)
	}
}

type SongLyricsResponse struct {
	Lyrics string `json:"lyrics"`
}

// @Summary Get the lyrics of a specific song
// @Tags API
// @Description Get song verses in partitions with pagination.
// @Param id path int true "ID of the song"
// @Param offset query int false "Offset for starting from a specific verse" default(0)
// @Param limit query int false "Maximum number of verses to retrieve" default(20)
// @Success 200 {object} SongLyricsResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /song/{id} [get]
func (s *Server) getSongLyricsHandler(w http.ResponseWriter, r *http.Request) {
	params, err := parseSongLyricsParams(r)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	lyrics, err := s.db.GetSongLyrics(params.ID, params.Offset, params.Limit)
	if err != nil {
		if err == postgres.SongNotFound {
			http.Error(w, "Song Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("error fetching song lyrics: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := SongLyricsResponse{lyrics}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("error encoding response: %v", err)
	}
}

type SongAddRequest struct {
	Song  string `json:"song"`
	Group string `json:"group"`
}

type SongAddResponse struct {
	ID int `json:"id"`
}

// @Summary Add a new song
// @Tags API
// @Description Add a new song to the library with the given title and group.
// @Param song body SongAddRequest true "Title and group"
// @Success 201 {object} SongAddResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /song [post]
func (s *Server) addSongHandler(w http.ResponseWriter, r *http.Request) {
	var req SongAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		log.Printf("Failed to decode JSON payload: %v", err)
		return
	}

	if req.Song == "" || req.Group == "" {
		http.Error(w, "Missing 'song' or 'group' query parameter", http.StatusBadRequest)
		return
	}

	songDetail, Err := rpc.GetSongDetail(req.Song, req.Group)
	if Err != nil {
		http.Error(w, Err.Status, Err.StatusCode)
		log.Println(Err.LogErr)
		return
	}

	song := &models.Song{
		SongInfo: models.SongInfo{
			Title:       req.Song,
			Group:       req.Group,
			ReleaseDate: songDetail.ReleaseDate,
			Link:        songDetail.Link,
		},
		Lyrics: songDetail.Text,
	}
	id, err := s.db.AddSong(song)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("failed to add song %v", err)
		return
	}

	resp := SongAddResponse{id}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("failed to write response: %v", err)
	}
}

// @Summary Update an existing song
// @Tags API
// @Description Update the details of an existing song, identified by its ID.
// The request body must contain the fields to be updated (e.g., song title, group, release date, or link).
// The song lyrics cannot be modified through this request.
// @Param id path int true "ID of the song to be updated"
// @Param song body models.SongInfo true "Updated song details"
// @Success 200 {object} models.SongInfo "Updated song details"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Song not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /song/{id} [put]
func (s *Server) updateSongHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	song := &models.SongInfo{ID: id}
	if err := json.NewDecoder(r.Body).Decode(song); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		log.Printf("failed to decode request body: %v", err)
		return
	}

	if err := s.db.UpdateSongInfo(song); err != nil {
		if err == postgres.SongNotFound {
			http.Error(w, "Song Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update song information", http.StatusInternalServerError)
			log.Printf("error updating song with id %d: %v", song.ID, err)
		}
	}
}

// @Summary Delete a song
// @Tags API
// @Description Delete a song from the library by its ID.
// Once deleted, the song and its details will be permanently removed from the database.
// @Param id path int true "ID of the song to be deleted"
// @Success 204 "Song successfully deleted"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /song/{id} [delete]
func (s *Server) deleteSongHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteSong(id); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("%v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
