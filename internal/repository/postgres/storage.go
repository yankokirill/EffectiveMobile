package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yankokirill/song-library/internal/models"
	"strings"
	"time"
)

type SongRepository interface {
	GetSongsInfo(hint *models.PaginationInfo) ([]models.SongInfo, error)
	GetGroupSongsInfo(group string, hint *models.PaginationInfo) ([]models.SongInfo, error)
	GetSongLyrics(id, offset, limit int) (string, error)

	AddSong(song *models.Song) (int, error)
	UpdateSongInfo(song *models.SongInfo) error

	DeleteSong(id int) error
	Clear() error

	Close()
}

type songRepo struct {
	pool *pgxpool.Pool
}

var SongNotFound = errors.New("song not found")

func NewSongRepository(databaseURL string) (SongRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	return &songRepo{pool: pool}, nil
}

func (sr *songRepo) GetSongsInfo(hint *models.PaginationInfo) ([]models.SongInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT * FROM get_songs_info($1, $2, $3)`
	rows, err := sr.pool.Query(ctx, query, hint.PrevSong, hint.PrevGroup, hint.Limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching songs: %w", err)
	}
	defer rows.Close()

	var songs []models.SongInfo
	for rows.Next() {
		var song models.SongInfo
		var date time.Time
		if err := rows.Scan(&song.ID, &song.Title, &song.Group, &date, &song.Link); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		song.ReleaseDate = date.Format("02.01.2006")
		songs = append(songs, song)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return songs, nil
}

func (sr *songRepo) GetGroupSongsInfo(group string, hint *models.PaginationInfo) ([]models.SongInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT * FROM get_group_songs_info($1, $2, $3)`
	rows, err := sr.pool.Query(ctx, query, group, hint.PrevSong, hint.Limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching songs: %w", err)
	}
	defer rows.Close()

	var songs []models.SongInfo
	for rows.Next() {
		var song models.SongInfo
		var date time.Time
		if err := rows.Scan(&song.ID, &song.Title, &song.Group, &date, &song.Link); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		song.ReleaseDate = date.Format("02.01.2006")
		songs = append(songs, song)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return songs, nil
}

func (sr *songRepo) GetSongLyrics(id, offset, limit int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT * FROM get_song_verses($1, $2, $3)`
	rows, err := sr.pool.Query(ctx, query, id, offset, limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var verses []string
	for rows.Next() {
		var verse string
		if err := rows.Scan(&verse); err != nil {
			return "", fmt.Errorf("error scanning row: %w", err)
		}
		verses = append(verses, verse)
	}

	if rows.Err() != nil {
		if pgErr, ok := rows.Err().(*pgconn.PgError); ok && pgErr.Code == "P0001" {
			return "", SongNotFound
		}
		return "", fmt.Errorf("error iterating rows: %w", rows.Err())
	}

	return strings.Join(verses, "\n\n"), nil
}

func (sr *songRepo) AddSong(song *models.Song) (int, error) {
	parsedDate, err := time.Parse("02.01.2006", song.ReleaseDate)
	if err != nil {
		return 0, err
	}
	verses := strings.Split(song.Lyrics, "\n\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int
	query := "SELECT add_song($1, $2, $3, $4, $5)"
	err = sr.pool.QueryRow(ctx, query, song.Title, song.Group, parsedDate, song.Link, verses).Scan(&id)
	if err != nil {

	}
	return id, err
}

func (sr *songRepo) UpdateSongInfo(song *models.SongInfo) error {
	parsedDate, err := time.Parse("02.01.2006", song.ReleaseDate)
	if err != nil {
		parsedDate = time.Time{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `CALL update_song_info($1, $2, $3, $4, $5)`
	_, err = sr.pool.Exec(ctx, query,
		song.ID,
		song.Title,
		song.Group,
		parsedDate,
		song.Link,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "P0001" {
			return SongNotFound
		}
	}
	return err
}

func (sr *songRepo) DeleteSong(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM songs WHERE id = $1`
	_, err := sr.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting song with id %d: %w", id, err)
	}

	return nil
}

func (sr *songRepo) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `TRUNCATE TABLE songs RESTART IDENTITY CASCADE;`
	_, err := sr.pool.Exec(ctx, query)
	return err
}

func (sr *songRepo) Close() {
	sr.pool.Close()
}
