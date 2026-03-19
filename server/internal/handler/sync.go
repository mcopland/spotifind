package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/repository"
	syncpkg "github.com/mcopland/spotifind/internal/sync"
)

type SyncHandler struct {
	syncService *syncpkg.Service
	syncRepo    *repository.SyncRepo
}

func NewSyncHandler(syncService *syncpkg.Service, syncRepo *repository.SyncRepo) *SyncHandler {
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
