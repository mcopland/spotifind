//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestPlaylistRepo_ListForUser(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "playlist_list_user_1")

	playlistRepo := repository.NewPlaylistRepo(testDB)
	pl, err := playlistRepo.Upsert(ctx, &models.Playlist{
		SpotifyID: "playlist_list_pl_1",
		Name:      "My Playlist",
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if err := playlistRepo.LinkToUser(ctx, user.ID, pl.ID); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}

	playlists, err := playlistRepo.ListForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListForUser failed: %v", err)
	}
	if len(playlists) != 1 {
		t.Fatalf("len(playlists): got %d, want 1", len(playlists))
	}
	if playlists[0].SpotifyID != pl.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", playlists[0].SpotifyID, pl.SpotifyID)
	}
}

func TestPlaylistRepo_GetStats(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "stats_user_1")

	artist := insertTestArtist(t, "stats_artist_1", nil)
	album := insertTestAlbum(t, "stats_album_1", 2020)
	track := insertTestTrack(t, "stats_track_1", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID); err != nil {
		t.Fatalf("LinkToUser track failed: %v", err)
	}

	albumRepo := repository.NewAlbumRepo(testDB)
	if err := albumRepo.LinkToUser(ctx, user.ID, album.ID); err != nil {
		t.Fatalf("LinkToUser album failed: %v", err)
	}

	artistRepo := repository.NewArtistRepo(testDB)
	if err := artistRepo.LinkToUser(ctx, user.ID, artist.ID); err != nil {
		t.Fatalf("LinkToUser artist failed: %v", err)
	}

	playlistRepo := repository.NewPlaylistRepo(testDB)
	pl, err := playlistRepo.Upsert(ctx, &models.Playlist{
		SpotifyID: "stats_playlist_1",
		Name:      "Stats Playlist",
	})
	if err != nil {
		t.Fatalf("Upsert playlist failed: %v", err)
	}
	if err := playlistRepo.LinkToUser(ctx, user.ID, pl.ID); err != nil {
		t.Fatalf("LinkToUser playlist failed: %v", err)
	}

	stats, err := playlistRepo.GetStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.Tracks != 1 {
		t.Errorf("Tracks: got %d, want 1", stats.Tracks)
	}
	if stats.Albums != 1 {
		t.Errorf("Albums: got %d, want 1", stats.Albums)
	}
	if stats.Artists != 1 {
		t.Errorf("Artists: got %d, want 1", stats.Artists)
	}
	if stats.Playlists != 1 {
		t.Errorf("Playlists: got %d, want 1", stats.Playlists)
	}
}

func TestPlaylistRepo_GetDistinctGenresForUser(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "genres_user_1")

	// Two artists share "rock"; only one "indie" and one "pop".
	// Distinct result should be ["indie", "pop", "rock"].
	artist1 := insertTestArtist(t, "genres_artist_1", []string{"rock", "indie"})
	artist2 := insertTestArtist(t, "genres_artist_2", []string{"rock", "pop"})

	album := insertTestAlbum(t, "genres_album_1", 2020)
	track1 := insertTestTrack(t, "genres_track_1", album.ID, false)
	track2 := insertTestTrack(t, "genres_track_2", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkArtist(ctx, track1.ID, artist1.ID); err != nil {
		t.Fatalf("LinkArtist track1 failed: %v", err)
	}
	if err := trackRepo.LinkArtist(ctx, track2.ID, artist2.ID); err != nil {
		t.Fatalf("LinkArtist track2 failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track1.ID); err != nil {
		t.Fatalf("LinkToUser track1 failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track2.ID); err != nil {
		t.Fatalf("LinkToUser track2 failed: %v", err)
	}

	playlistRepo := repository.NewPlaylistRepo(testDB)
	genres, err := playlistRepo.GetDistinctGenresForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetDistinctGenresForUser failed: %v", err)
	}
	if len(genres) != 3 {
		t.Errorf("genres count: got %d, want 3 (got %v)", len(genres), genres)
	}
	expected := []string{"indie", "pop", "rock"}
	for i, want := range expected {
		if i >= len(genres) {
			break
		}
		if genres[i] != want {
			t.Errorf("genres[%d]: got %q, want %q", i, genres[i], want)
		}
	}
}
