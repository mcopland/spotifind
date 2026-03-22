//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func TestUserRepo_UpsertCreates(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepo(testDB)

	u := &models.User{
		SpotifyID:      "user_create_1",
		DisplayName:    "Test User",
		Email:          "test@example.com",
		TokenExpiresAt: time.Now().Add(time.Hour),
	}
	got, err := repo.Upsert(ctx, u)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if got.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if got.SpotifyID != u.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", got.SpotifyID, u.SpotifyID)
	}
	if got.DisplayName != u.DisplayName {
		t.Errorf("DisplayName: got %q, want %q", got.DisplayName, u.DisplayName)
	}
}

func TestUserRepo_UpsertUpdates(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepo(testDB)

	u := &models.User{
		SpotifyID:      "user_update_1",
		DisplayName:    "Original Name",
		TokenExpiresAt: time.Now().Add(time.Hour),
	}
	_, err := repo.Upsert(ctx, u)
	if err != nil {
		t.Fatalf("first Upsert failed: %v", err)
	}

	u.DisplayName = "Updated Name"
	got, err := repo.Upsert(ctx, u)
	if err != nil {
		t.Fatalf("second Upsert failed: %v", err)
	}
	if got.DisplayName != "Updated Name" {
		t.Errorf("DisplayName: got %q, want %q", got.DisplayName, "Updated Name")
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepo(testDB)

	u := &models.User{
		SpotifyID:      "user_getbyid_1",
		DisplayName:    "Get By ID User",
		TokenExpiresAt: time.Now().Add(time.Hour),
	}
	created, err := repo.Upsert(ctx, u)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: got %d, want %d", got.ID, created.ID)
	}
	if got.DisplayName != u.DisplayName {
		t.Errorf("DisplayName: got %q, want %q", got.DisplayName, u.DisplayName)
	}
}

func TestUserRepo_GetBySpotifyID(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepo(testDB)

	u := &models.User{
		SpotifyID:      "user_getbyspotify_1",
		DisplayName:    "Get By SpotifyID User",
		TokenExpiresAt: time.Now().Add(time.Hour),
	}
	_, err := repo.Upsert(ctx, u)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	got, err := repo.GetBySpotifyID(ctx, u.SpotifyID)
	if err != nil {
		t.Fatalf("GetBySpotifyID failed: %v", err)
	}
	if got.SpotifyID != u.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", got.SpotifyID, u.SpotifyID)
	}
}
