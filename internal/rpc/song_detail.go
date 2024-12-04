package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/yankokirill/song-library/internal/models"
	"net/http"
	"net/url"
	"time"
)

var externalApiURL string

type HttpError struct {
	StatusCode int
	Status     string
	LogErr     error
}

func SetExternalApiURL(url string) {
	externalApiURL = url
}

func GetSongDetail(songTitle, groupName string) (*models.SongDetail, *HttpError) {
	Err := &HttpError{StatusCode: http.StatusOK}

	query := url.Values{}
	query.Set("song", songTitle)
	query.Set("group", groupName)

	client := &http.Client{Timeout: 5 * time.Second}
	reqURL := fmt.Sprintf("%s/info?%s", externalApiURL, query.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		Err.Status = "Internal Server Error"
		Err.StatusCode = http.StatusInternalServerError
		Err.LogErr = fmt.Errorf("error preparing request to external api: %w", err)
		return nil, Err
	}
	req.Header.Set("Accept", "application/json")

	respRPC, err := client.Do(req)
	if err != nil {
		Err.Status = "Internal Server Error"
		Err.StatusCode = http.StatusInternalServerError
		Err.LogErr = fmt.Errorf("error sending request to external api: %w", err)
		return nil, Err
	}
	defer respRPC.Body.Close()

	if respRPC.StatusCode != http.StatusOK {
		Err.Status = respRPC.Status
		Err.StatusCode = respRPC.StatusCode
		Err.LogErr = fmt.Errorf("failed rpc.GetSongInfo(%s, %s), (status code: %d, status: %s)",
			songTitle, groupName, Err.StatusCode, Err.Status)
		return nil, Err
	}

	var songDetail models.SongDetail
	if err := json.NewDecoder(respRPC.Body).Decode(&songDetail); err != nil {
		Err.Status = "Internal Server Error"
		Err.StatusCode = http.StatusInternalServerError
		Err.LogErr = fmt.Errorf("error decoding json from external api: %w", err)
		return nil, Err
	}
	return &songDetail, nil
}
