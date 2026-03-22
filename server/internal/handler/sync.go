package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// SyncStarter is satisfied by sync.Service.
type SyncStarter interface {
	StartSync(userID int64) (int64, error)
}

// SyncStatusGetter is satisfied by repository.SyncRepo.
type SyncStatusGetter interface {
	GetLatestForUser(ctx context.Context, userID int64) (*models.SyncJob, error)
}

type SyncHandler struct {
	syncService SyncStarter
	syncRepo    SyncStatusGetter
}

func NewSyncHandler(syncService SyncStarter, syncRepo SyncStatusGetter) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		syncRepo:    syncRepo,
	}
}

func (h *SyncHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	jobID, err := h.syncService.StartSync(userID)
	if err != nil {
		http.Error(w, "failed to start sync", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]int64{"job_id": jobID})
}

func (h *SyncHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	job, err := h.syncRepo.GetLatestForUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "none"})
			return
		}
		http.Error(w, "failed to get sync status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
