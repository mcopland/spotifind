//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestArtistRepo_ListForUser_ReturnsInserted(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "artist_list_user_1")
	artist := insertTestArtist(t, "artist_list_ar_1", []string{"jazz"})

	artistRepo := repository.NewArtistRepo(testDB)
	if err := artistRepo.LinkToUser(ctx, user.ID, artist.ID); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}

	result, err := artistRepo.ListForUser(ctx, user.ID, models.ArtistFilters{})
	if err != nil {
		t.Fatalf("ListForUser failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Items: got %d, want 1", len(result.Items))
	}
	if result.Items[0].SpotifyID != artist.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, artist.SpotifyID)
	}
}

func TestArtistRepo_ListForUser_GenreFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "artist_genre_user_1")
	artistRock := insertTestArtist(t, "artist_genre_ar_rock", []string{"rock"})
	artistJazz := insertTestArtist(t, "artist_genre_ar_jazz", []string{"jazz"})

	artistRepo := repository.NewArtistRepo(testDB)
	if err := artistRepo.LinkToUser(ctx, user.ID, artistRock.ID); err != nil {
		t.Fatalf("LinkToUser artistRock failed: %v", err)
	}
	if err := artistRepo.LinkToUser(ctx, user.ID, artistJazz.ID); err != nil {
		t.Fatalf("LinkToUser artistJazz failed: %v", err)
	}

	result, err := artistRepo.ListForUser(ctx, user.ID, models.ArtistFilters{Genres: []string{"rock"}})
	if err != nil {
		t.Fatalf("ListForUser with genre filter failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != artistRock.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, artistRock.SpotifyID)
	}
}
