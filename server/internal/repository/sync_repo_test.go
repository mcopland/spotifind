//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/repository"
)

func TestSyncRepo_CreateAndGetLatest(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "sync_latest_user_1")

	syncRepo := repository.NewSyncRepo(testDB)
	job, err := syncRepo.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if job.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if job.Status != "pending" {
		t.Errorf("Status: got %q, want %q", job.Status, "pending")
	}

	latest, err := syncRepo.GetLatestForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetLatestForUser failed: %v", err)
	}
	if latest.ID != job.ID {
		t.Errorf("ID: got %d, want %d", latest.ID, job.ID)
	}
}

func TestSyncRepo_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "sync_status_user_1")

	syncRepo := repository.NewSyncRepo(testDB)
	job, err := syncRepo.Create(ctx, user.ID)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := syncRepo.UpdateStatus(ctx, job.ID, "running"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	latest, err := syncRepo.GetLatestForUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetLatestForUser failed: %v", err)
	}
	if latest.Status != "running" {
		t.Errorf("Status: got %q, want %q", latest.Status, "running")
	}
}
