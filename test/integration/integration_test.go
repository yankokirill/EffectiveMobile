package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	. "github.com/yankokirill/song-library/internal/delivery/http"
	"github.com/yankokirill/song-library/internal/models"
	"github.com/yankokirill/song-library/internal/repository/migrations"
	. "github.com/yankokirill/song-library/internal/repository/postgres"
	"github.com/yankokirill/song-library/internal/rpc"
	"github.com/yankokirill/song-library/test/mock"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
)

var baseURL = "http://localhost:8080/library"
var dbURL = "postgres://user:password@localhost:5432/test_db?sslmode=disable"
var repo SongRepository

type AddRequest struct {
	Song  string `json:"song"`
	Group string `json:"group"`
}

func TestMain(m *testing.M) {
	mockServer := mock.NewExternalApiServer()
	defer mockServer.Close()
	rpc.SetExternalApiURL(mockServer.URL)

	ctr, err := postgres.Run(context.Background(),
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	dbURL, err = ctr.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get database URL: %v", err)
	}

	repo, err = NewSongRepository(dbURL)
	if err != nil {
		log.Fatalf("failed to start postgresql database: %v", err)
	}
	defer repo.Close()

	if err := migrations.Up("file://../../migrations", dbURL); err != nil {
		log.Fatalf("%v", err)
	}

	server := NewServer(repo, ":8080")
	handler := httptest.NewServer(server.Routes())
	defer handler.Close()
	baseURL = handler.URL + "/library"

	code := m.Run()

	os.Exit(code)
}

func AddSong(t *testing.T, req *AddRequest, statusCode int) {
	t.Helper()
	songJSON, err := json.Marshal(req)
	require.NoError(t, err, "Failed to marshal song")

	resp, err := http.Post(baseURL+"/song", "application/json", bytes.NewBuffer(songJSON))
	require.NoError(t, err, "Failed to make POST request")
	defer resp.Body.Close()

	require.Equal(t, statusCode, resp.StatusCode, "Expected status %d, got %d", statusCode, resp.StatusCode)
}

func UpdateSong(t *testing.T, reqJSON string, id, statusCode int) {
	t.Helper()
	url := fmt.Sprintf("%s/song/%d", baseURL, id)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(reqJSON)))
	require.NoError(t, err, "Failed to prepare PUT request")

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to make PUT request")
	defer resp.Body.Close()

	require.Equal(t, statusCode, resp.StatusCode, "Expected status %d, got %d", statusCode, resp.StatusCode)
}

func GetLyrics(t *testing.T, offset, limit, id, statusCode int) string {
	t.Helper()
	url := fmt.Sprintf("%s/song/%d?offset=%d&limit=%d", baseURL, id, offset, limit)
	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to make GET request")
	defer resp.Body.Close()

	require.Equal(t, statusCode, resp.StatusCode, "Expected status %d, got %d", statusCode, resp.StatusCode)
	if statusCode != http.StatusOK {
		return ""
	}

	data := struct {
		Lyrics string `json:"lyrics"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err, "Failed to decode response")
	return data.Lyrics
}

func DeleteSong(t *testing.T, id int) {
	t.Helper()
	url := fmt.Sprintf("%s/song/%d", baseURL, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err, "Failed to prepare DELETE request")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to make DELETE request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status 204, got %d", resp.StatusCode)
}

func GetQuery(t *testing.T, url string) (songs []models.SongInfo) {
	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to make GET request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200, got %d", resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&songs)
	require.NoError(t, err, "Failed to decode response")
	return
}

func GetSongs(t *testing.T) []models.SongInfo {
	return GetQuery(t, baseURL+"/songs")
}

func GetSongsPagination(t *testing.T, song, group string, limit int) []models.SongInfo {
	t.Helper()

	base, err := url.Parse(fmt.Sprintf("%s/songs", baseURL))
	require.NoError(t, err, "Failed to parse base URL")

	query := base.Query()
	if song != "" {
		query.Set("prevSong", song)
	}
	if group != "" {
		query.Set("prevGroup", group)
	}
	query.Set("limit", strconv.Itoa(limit))
	base.RawQuery = query.Encode()

	return GetQuery(t, base.String())
}

func GetGroupSongsPagination(t *testing.T, group, song string, limit int) []models.SongInfo {
	t.Helper()

	base, err := url.Parse(fmt.Sprintf("%s/songs/%s", baseURL, group))
	require.NoError(t, err, "Failed to parse base URL")

	query := base.Query()
	if song != "" {
		query.Set("prevSong", song)
	}
	query.Set("limit", strconv.Itoa(limit))
	base.RawQuery = query.Encode()

	return GetQuery(t, base.String())
}

func TestAddSong(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Supermassive Black Hole",
		Group: "Muse",
	}
	AddSong(t, &songName, http.StatusCreated)

	songs := GetSongs(t)
	require.Equal(t, 1, len(songs))
	require.Equal(t, "16.07.2006", songs[0].ReleaseDate)
	require.Equal(t, "https://www.youtube.com/watch?v=Xsp3_a-PMTw", songs[0].Link)
}

func TestUpdateSong_Date(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Supermassive Black Hole",
		Group: "Muse",
	}
	AddSong(t, &songName, http.StatusCreated)

	UpdateSong(t, `{"releaseDate": "19.06.2006"}`, 1, http.StatusOK)
	expected := models.SongInfo{
		ID:          1,
		Title:       "Supermassive Black Hole",
		Group:       "Muse",
		ReleaseDate: "19.06.2006",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}
	songs := GetSongs(t)
	require.Equal(t, 1, len(songs))
	require.Equal(t, expected, songs[0])
}

func TestUpdateSong_Group(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Supermassive Black Hole",
		Group: "Muse",
	}
	AddSong(t, &songName, http.StatusCreated)

	UpdateSong(t, `{"group": "Sus"}`, 1, http.StatusOK)
	expected := models.SongInfo{
		ID:          1,
		Title:       "Supermassive Black Hole",
		Group:       "Sus",
		ReleaseDate: "16.07.2006",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}
	songs := GetSongs(t)
	require.Equal(t, 1, len(songs))
	require.Equal(t, expected, songs[0])
}

func TestUpdateSong_NotFound(t *testing.T) {
	UpdateSong(t, `{"group": "Sus"}`, 1, http.StatusNotFound)
}

func TestDeleteSong(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Supermassive Black Hole",
		Group: "Muse",
	}
	AddSong(t, &songName, http.StatusCreated)

	songName = AddRequest{
		Song:  "Yellow",
		Group: "Coldplay",
	}
	AddSong(t, &songName, http.StatusCreated)

	DeleteSong(t, 1)
	expected := models.SongInfo{
		ID:          2,
		Title:       "Yellow",
		Group:       "Coldplay",
		ReleaseDate: "26.06.2000",
		Link:        "https://www.youtube.com/watch?v=yKNxeF4KMsY",
	}
	songs := GetSongs(t)
	require.Equal(t, 1, len(songs))
	require.Equal(t, expected, songs[0])
}

func TestGetSongLyrics(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Supermassive Black Hole",
		Group: "Muse",
	}
	AddSong(t, &songName, http.StatusCreated)

	lyrics := GetLyrics(t, 2, 1, 1, http.StatusOK)
	expected := "Glaciers melting in the dead of night (ooh)\nAnd the superstars sucked into the supermassive (you set my soul alight)\nGlaciers melting in the dead of night\nAnd the superstars sucked into the (you set my soul)\n(Into the supermassive)"
	require.Equal(t, expected, lyrics)
}

func TestGetSongLyrics_Sample(t *testing.T) {
	defer repo.Clear()

	songName := AddRequest{
		Song:  "Sample",
		Group: "Sample",
	}
	AddSong(t, &songName, http.StatusCreated)

	lyrics := GetLyrics(t, 2, 3, 1, http.StatusOK)
	expected := "3\n\n4\n\n5"
	require.Equal(t, expected, lyrics)
}

func TestGetSongLyrics_NotFound(t *testing.T) {
	GetLyrics(t, 2, 3, 1, http.StatusNotFound)
}

func PreparePaginationCase(t *testing.T) {
	t.Helper()
	for groupNum := range 3 {
		for songNum := range 4 {
			songName := AddRequest{
				Song:  fmt.Sprintf("%d", songNum+1),
				Group: fmt.Sprintf("Group %d", groupNum+1),
			}
			AddSong(t, &songName, http.StatusCreated)
		}
	}
}

func TestPreparePagination(t *testing.T) {
	defer repo.Clear()
	PreparePaginationCase(t)

	songs := GetSongsPagination(t, "", "", 4)
	require.Equal(t, 4, len(songs))

	expected := make([]models.SongInfo, 4)
	expected[0] = models.SongInfo{
		ID:          1,
		Title:       "1",
		Group:       "Group 1",
		ReleaseDate: "01.01.2025",
	}
	expected[1] = models.SongInfo{
		ID:          5,
		Title:       "1",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	expected[2] = models.SongInfo{
		ID:          9,
		Title:       "1",
		Group:       "Group 3",
		ReleaseDate: "01.01.2025",
	}
	expected[3] = models.SongInfo{
		ID:          2,
		Title:       "2",
		Group:       "Group 1",
		ReleaseDate: "01.01.2025",
	}
	require.Equal(t, expected, songs)
}

func TestGetSongsPagination_All(t *testing.T) {
	defer repo.Clear()
	PreparePaginationCase(t)

	songs := GetSongsPagination(t, "", "", 4)
	require.Equal(t, 4, len(songs))

	expected := make([]models.SongInfo, 4)
	expected[0] = models.SongInfo{
		ID:          1,
		Title:       "1",
		Group:       "Group 1",
		ReleaseDate: "01.01.2025",
	}
	expected[1] = models.SongInfo{
		ID:          5,
		Title:       "1",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	expected[2] = models.SongInfo{
		ID:          9,
		Title:       "1",
		Group:       "Group 3",
		ReleaseDate: "01.01.2025",
	}
	expected[3] = models.SongInfo{
		ID:          2,
		Title:       "2",
		Group:       "Group 1",
		ReleaseDate: "01.01.2025",
	}
	require.Equal(t, expected, songs)

	songs = GetSongsPagination(t, "2", "Group 1", 1)
	expected = expected[0:1]
	expected[0] = models.SongInfo{
		ID:          6,
		Title:       "2",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	require.Equal(t, expected, songs)
}

func TestGetSongsPagination_Group(t *testing.T) {
	defer repo.Clear()
	PreparePaginationCase(t)

	songs := GetGroupSongsPagination(t, "Group 2", "", 2)
	expected := make([]models.SongInfo, 2)
	expected[0] = models.SongInfo{
		ID:          5,
		Title:       "1",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	expected[1] = models.SongInfo{
		ID:          6,
		Title:       "2",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	require.Equal(t, expected, songs)

	songs = GetGroupSongsPagination(t, "Group 2", "2", 3)
	expected[0] = models.SongInfo{
		ID:          7,
		Title:       "3",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	expected[1] = models.SongInfo{
		ID:          8,
		Title:       "4",
		Group:       "Group 2",
		ReleaseDate: "01.01.2025",
	}
	require.Equal(t, expected, songs)
}
