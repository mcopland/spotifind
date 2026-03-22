//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://spotifind:spotifind@localhost:5433/spotifind_test"
	}

	var err error
	testDB, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	mg, err := migrate.New("file://../database/migrations", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		os.Exit(1)
	}

	if err := mg.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}

	srcErr, dbErr := mg.Close()
	if srcErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close migrate source: %v\n", srcErr)
	}
	if dbErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close migrate db: %v\n", dbErr)
	}

	truncateAll()

	os.Exit(m.Run())
}

func truncateAll() {
	_, err := testDB.Exec(context.Background(), `
		TRUNCATE
			user_top_artists, user_top_tracks, user_recently_played,
			playlist_tracks, user_playlists, user_followed_artists,
			user_saved_albums, user_saved_tracks, sync_jobs,
			track_artists, album_artists,
			tracks, playlists, albums, artists, users
		CASCADE`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to truncate tables: %v\n", err)
		os.Exit(1)
	}
}

func insertTestUser(t *testing.T, spotifyID string) *models.User {
	t.Helper()
	repo := repository.NewUserRepo(testDB)
	u, err := repo.Upsert(context.Background(), &models.User{
		SpotifyID:      spotifyID,
		DisplayName:    "Test User",
		TokenExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to insert user %q: %v", spotifyID, err)
	}
	return u
}

func insertTestArtist(t *testing.T, spotifyID string, genres []string) *models.Artist {
	t.Helper()
	if genres == nil {
		genres = []string{}
	}
	repo := repository.NewArtistRepo(testDB)
	a, err := repo.Upsert(context.Background(), &models.Artist{
		SpotifyID: spotifyID,
		Name:      "Artist " + spotifyID,
		Genres:    genres,
	})
	if err != nil {
		t.Fatalf("failed to insert artist %q: %v", spotifyID, err)
	}
	return a
}

func insertTestAlbum(t *testing.T, spotifyID string, releaseYear int) *models.Album {
	t.Helper()
	repo := repository.NewAlbumRepo(testDB)
	al, err := repo.Upsert(context.Background(), &models.Album{
		SpotifyID:   spotifyID,
		Name:        "Album " + spotifyID,
		AlbumType:   "album",
		ReleaseYear: releaseYear,
	})
	if err != nil {
		t.Fatalf("failed to insert album %q: %v", spotifyID, err)
	}
	return al
}

func insertTestTrack(t *testing.T, spotifyID string, albumID int64, explicit bool) *models.Track {
	t.Helper()
	repo := repository.NewTrackRepo(testDB)
	tr, err := repo.Upsert(context.Background(), &models.Track{
		SpotifyID: spotifyID,
		Name:      "Track " + spotifyID,
		AlbumID:   &albumID,
		Explicit:  explicit,
	})
	if err != nil {
		t.Fatalf("failed to insert track %q: %v", spotifyID, err)
	}
	return tr
}
