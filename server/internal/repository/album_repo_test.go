//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestAlbumRepo_ListForUser_ReturnsInserted(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "album_list_user_1")
	album := insertTestAlbum(t, "album_list_al_1", 2020)

	albumRepo := repository.NewAlbumRepo(testDB)
	if err := albumRepo.LinkToUser(ctx, user.ID, album.ID); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}

	result, err := albumRepo.ListForUser(ctx, user.ID, models.AlbumFilters{})
	if err != nil {
		t.Fatalf("ListForUser failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Items: got %d, want 1", len(result.Items))
	}
	if result.Items[0].SpotifyID != album.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, album.SpotifyID)
	}
}

func TestAlbumRepo_ListForUser_YearRange(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "album_year_user_1")
	al2018 := insertTestAlbum(t, "album_year_al_2018", 2018)
	al2020 := insertTestAlbum(t, "album_year_al_2020", 2020)
	al2022 := insertTestAlbum(t, "album_year_al_2022", 2022)

	albumRepo := repository.NewAlbumRepo(testDB)
	for _, al := range []*models.Album{al2018, al2020, al2022} {
		if err := albumRepo.LinkToUser(ctx, user.ID, al.ID); err != nil {
			t.Fatalf("LinkToUser failed for album %q: %v", al.SpotifyID, err)
		}
	}

	yearMin, yearMax := 2019, 2021
	result, err := albumRepo.ListForUser(ctx, user.ID, models.AlbumFilters{YearMin: &yearMin, YearMax: &yearMax})
	if err != nil {
		t.Fatalf("ListForUser with year range failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].ReleaseYear != 2020 {
		t.Errorf("ReleaseYear: got %d, want 2020", result.Items[0].ReleaseYear)
	}
}
