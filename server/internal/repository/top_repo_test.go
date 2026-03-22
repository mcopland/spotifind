//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestTopRepo_UpsertAndListTopTracks(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "top-user-1")
	al := insertTestAlbum(t, "top-album-1", 2020)
	tr1 := insertTestTrack(t, "top-track-1", al.ID, false)
	tr2 := insertTestTrack(t, "top-track-2", al.ID, false)

	repo := repository.NewTopRepo(testDB)
	ctx := context.Background()

	if err := repo.DeleteTopTracksForUser(ctx, u.ID, "short_term"); err != nil {
		t.Fatalf("DeleteTopTracksForUser: %v", err)
	}
	if err := repo.UpsertTopTrack(ctx, u.ID, tr1.ID, 1, "short_term"); err != nil {
		t.Fatalf("UpsertTopTrack rank 1: %v", err)
	}
	if err := repo.UpsertTopTrack(ctx, u.ID, tr2.ID, 2, "short_term"); err != nil {
		t.Fatalf("UpsertTopTrack rank 2: %v", err)
	}

	result, err := repo.ListTopTracksForUser(ctx, u.ID, models.TopFilters{TimeRange: "short_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopTracksForUser: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Total: want 2, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("Items len: want 2, got %d", len(result.Items))
	}
	if result.Items[0].Rank != 1 {
		t.Errorf("first item rank: want 1, got %d", result.Items[0].Rank)
	}
	if result.Items[1].Rank != 2 {
		t.Errorf("second item rank: want 2, got %d", result.Items[1].Rank)
	}
}

func TestTopRepo_ListTopTracks_TimeRangeFilter(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "top-user-2")
	al := insertTestAlbum(t, "top-album-2", 2021)
	tr1 := insertTestTrack(t, "top-track-3", al.ID, false)
	tr2 := insertTestTrack(t, "top-track-4", al.ID, false)

	repo := repository.NewTopRepo(testDB)
	ctx := context.Background()

	if err := repo.UpsertTopTrack(ctx, u.ID, tr1.ID, 1, "short_term"); err != nil {
		t.Fatalf("UpsertTopTrack short_term: %v", err)
	}
	if err := repo.UpsertTopTrack(ctx, u.ID, tr2.ID, 1, "medium_term"); err != nil {
		t.Fatalf("UpsertTopTrack medium_term: %v", err)
	}

	short, err := repo.ListTopTracksForUser(ctx, u.ID, models.TopFilters{TimeRange: "short_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopTracksForUser short_term: %v", err)
	}
	if short.Total != 1 {
		t.Errorf("short_term Total: want 1, got %d", short.Total)
	}
	if short.Items[0].SpotifyID != "top-track-3" {
		t.Errorf("short_term item: want top-track-3, got %q", short.Items[0].SpotifyID)
	}

	medium, err := repo.ListTopTracksForUser(ctx, u.ID, models.TopFilters{TimeRange: "medium_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopTracksForUser medium_term: %v", err)
	}
	if medium.Total != 1 {
		t.Errorf("medium_term Total: want 1, got %d", medium.Total)
	}
	if medium.Items[0].SpotifyID != "top-track-4" {
		t.Errorf("medium_term item: want top-track-4, got %q", medium.Items[0].SpotifyID)
	}
}

func TestTopRepo_UpsertAndListTopArtists(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "top-user-3")
	ar1 := insertTestArtist(t, "top-artist-1", nil)
	ar2 := insertTestArtist(t, "top-artist-2", nil)

	repo := repository.NewTopRepo(testDB)
	ctx := context.Background()

	if err := repo.DeleteTopArtistsForUser(ctx, u.ID, "short_term"); err != nil {
		t.Fatalf("DeleteTopArtistsForUser: %v", err)
	}
	if err := repo.UpsertTopArtist(ctx, u.ID, ar1.ID, 1, "short_term"); err != nil {
		t.Fatalf("UpsertTopArtist rank 1: %v", err)
	}
	if err := repo.UpsertTopArtist(ctx, u.ID, ar2.ID, 2, "short_term"); err != nil {
		t.Fatalf("UpsertTopArtist rank 2: %v", err)
	}

	result, err := repo.ListTopArtistsForUser(ctx, u.ID, models.TopFilters{TimeRange: "short_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopArtistsForUser: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Total: want 2, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("Items len: want 2, got %d", len(result.Items))
	}
	if result.Items[0].Rank != 1 {
		t.Errorf("first item rank: want 1, got %d", result.Items[0].Rank)
	}
	if result.Items[1].Rank != 2 {
		t.Errorf("second item rank: want 2, got %d", result.Items[1].Rank)
	}
}

func TestTopRepo_ListTopArtists_TimeRangeFilter(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "top-user-4")
	ar1 := insertTestArtist(t, "top-artist-3", nil)
	ar2 := insertTestArtist(t, "top-artist-4", nil)

	repo := repository.NewTopRepo(testDB)
	ctx := context.Background()

	if err := repo.UpsertTopArtist(ctx, u.ID, ar1.ID, 1, "short_term"); err != nil {
		t.Fatalf("UpsertTopArtist short_term: %v", err)
	}
	if err := repo.UpsertTopArtist(ctx, u.ID, ar2.ID, 1, "medium_term"); err != nil {
		t.Fatalf("UpsertTopArtist medium_term: %v", err)
	}

	short, err := repo.ListTopArtistsForUser(ctx, u.ID, models.TopFilters{TimeRange: "short_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopArtistsForUser short_term: %v", err)
	}
	if short.Total != 1 {
		t.Errorf("short_term Total: want 1, got %d", short.Total)
	}
	if short.Items[0].SpotifyID != "top-artist-3" {
		t.Errorf("short_term item: want top-artist-3, got %q", short.Items[0].SpotifyID)
	}

	medium, err := repo.ListTopArtistsForUser(ctx, u.ID, models.TopFilters{TimeRange: "medium_term", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListTopArtistsForUser medium_term: %v", err)
	}
	if medium.Total != 1 {
		t.Errorf("medium_term Total: want 1, got %d", medium.Total)
	}
	if medium.Items[0].SpotifyID != "top-artist-4" {
		t.Errorf("medium_term item: want top-artist-4, got %q", medium.Items[0].SpotifyID)
	}
}
