//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestRecentlyPlayedRepo_Upsert_And_ListForUser(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "rp-user-1")
	al := insertTestAlbum(t, "rp-album-1", 2020)
	tr1 := insertTestTrack(t, "rp-track-1", al.ID, false)
	tr2 := insertTestTrack(t, "rp-track-2", al.ID, false)

	repo := repository.NewRecentlyPlayedRepo(testDB)
	ctx := context.Background()

	t1 := time.Now().UTC().Truncate(time.Millisecond)
	t2 := t1.Add(-time.Hour)

	if err := repo.Upsert(ctx, u.ID, tr1.ID, t1); err != nil {
		t.Fatalf("Upsert tr1: %v", err)
	}
	if err := repo.Upsert(ctx, u.ID, tr2.ID, t2); err != nil {
		t.Fatalf("Upsert tr2: %v", err)
	}

	result, err := repo.ListForUser(ctx, u.ID, models.RecentlyPlayedFilters{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Total: want 2, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("Items len: want 2, got %d", len(result.Items))
	}
	// Results must be ordered by played_at DESC: tr1 (most recent) first.
	if result.Items[0].SpotifyID != "rp-track-1" {
		t.Errorf("first item: want rp-track-1, got %q", result.Items[0].SpotifyID)
	}
	if result.Items[1].SpotifyID != "rp-track-2" {
		t.Errorf("second item: want rp-track-2, got %q", result.Items[1].SpotifyID)
	}
}

func TestRecentlyPlayedRepo_Upsert_Deduplicates(t *testing.T) {
	truncateAll()

	u := insertTestUser(t, "rp-user-2")
	al := insertTestAlbum(t, "rp-album-2", 2021)
	tr := insertTestTrack(t, "rp-track-3", al.ID, false)

	repo := repository.NewRecentlyPlayedRepo(testDB)
	ctx := context.Background()

	playedAt := time.Now().UTC().Truncate(time.Millisecond)

	if err := repo.Upsert(ctx, u.ID, tr.ID, playedAt); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	if err := repo.Upsert(ctx, u.ID, tr.ID, playedAt); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	result, err := repo.ListForUser(ctx, u.ID, models.RecentlyPlayedFilters{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: want 1 after dedup, got %d", result.Total)
	}
}
