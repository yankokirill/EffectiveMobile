package mock_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/yankokirill/song-library/internal/models"
	"github.com/yankokirill/song-library/test/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	server = mock.NewExternalApiServer()
	defer server.Close()

	code := m.Run()

	os.Exit(code)
}

func SendQuery(t *testing.T, songInfo *mock.SongInfo) *http.Response {
	t.Helper()
	reqURL, err := url.Parse(server.URL + "/info")
	require.NoError(t, err, "Failed to parse server URL")

	query := reqURL.Query()
	query.Set("group", songInfo.Group)
	query.Set("song", songInfo.Title)
	reqURL.RawQuery = query.Encode()

	resp, err := http.Get(reqURL.String())
	require.NoError(t, err, "Request to mock server failed")
	return resp
}

func TestQuery_Success(t *testing.T) {
	songInfo := &mock.SongInfo{
		Title: "Supermassive Black Hole",
		Group: "Muse",
	}
	expected := models.SongDetail{
		ReleaseDate: "16.07.2006",
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight\n\nGlaciers melting in the dead of night (ooh)\nAnd the superstars sucked into the supermassive (you set my soul alight)\nGlaciers melting in the dead of night\nAnd the superstars sucked into the (you set my soul)\n(Into the supermassive)\n\nI thought I was a fool for no one\nOoh baby, I'm a fool for you\nYou're the queen of the superficial\nAnd how long before you tell the truth?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight\n\nGlaciers melting in the dead of night (ooh)\nAnd the superstars sucked into the supermassive (you set my soul alight)\nGlaciers melting in the dead of night\nAnd the superstars sucked into the (you set my soul)\n(Into the supermassive)\n\nSupermassive black hole\nSupermassive black hole\nSupermassive black hole\nSupermassive black hole\n\nGlaciers melting in the dead of night\nAnd the superstars sucked into the supermassive\nGlaciers melting in the dead of night\nAnd the superstars sucked into the supermassive\nGlaciers melting in the dead of night (ooh)\nAnd the superstars sucked into the supermassive (you set my soul alight)\nGlaciers melting in the dead of night\nAnd the superstars sucked into the (you set my soul)\n(Into the supermassive)\n\nSupermassive black hole\nSupermassive black hole\nSupermassive black hole\nSupermassive black hole",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}

	resp := SendQuery(t, songInfo)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code")

	var result models.SongDetail
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode JSON response")

	require.Equal(t, expected, result, "Unexpected result")
}

func TestQuery_NotFound(t *testing.T) {
	songInfo := &mock.SongInfo{
		Title: "Unknown Title",
		Group: "Muse",
	}

	resp := SendQuery(t, songInfo)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode, "Unexpected status code")
}

func TestQuery_InternalError(t *testing.T) {
	songInfo := &mock.SongInfo{
		Title: "Faint",
		Group: "Linkin Park",
	}

	resp := SendQuery(t, songInfo)
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Unexpected status code")
}
