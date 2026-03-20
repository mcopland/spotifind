package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// RecentlyPlayedLister is satisfied by repository.RecentlyPlayedRepo.
type RecentlyPlayedLister interface {
	ListForUser(ctx context.Context, userID int64, f models.RecentlyPlayedFilters) (*models.PaginatedResult[models.RecentlyPlayedTrack], error)
}

type RecentlyPlayedHandler struct {
	repo RecentlyPlayedLister
}

func NewRecentlyPlayedHandler(repo RecentlyPlayedLister) *RecentlyPlayedHandler {
	return &RecentlyPlayedHandler{repo: repo}
}

func (h *RecentlyPlayedHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	f := models.RecentlyPlayedFilters{}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Page = n
		}
	}
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.PageSize = n
		}
	}

	result, err := h.repo.ListForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list recently played", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
