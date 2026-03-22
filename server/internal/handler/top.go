package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// TopLister is satisfied by repository.TopRepo.
type TopLister interface {
	ListTopTracksForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopTrack], error)
	ListTopArtistsForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopArtist], error)
}

type TopHandler struct {
	repo TopLister
}

func NewTopHandler(repo TopLister) *TopHandler {
	return &TopHandler{repo: repo}
}

func (h *TopHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	f := parseTopFilters(r)
	result, err := h.repo.ListTopTracksForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list top tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *TopHandler) ListArtists(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	f := parseTopFilters(r)
	result, err := h.repo.ListTopArtistsForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list top artists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func parseTopFilters(r *http.Request) models.TopFilters {
	q := r.URL.Query()
	f := models.TopFilters{
		TimeRange: q.Get("time_range"),
	}
	if f.TimeRange == "" {
		f.TimeRange = "short_term"
	}
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
	return f
}
