package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// ArtistLister is satisfied by repository.ArtistRepo.
type ArtistLister interface {
	ListForUser(ctx context.Context, userID int64, f models.ArtistFilters) (*models.PaginatedResult[models.Artist], error)
}

type ArtistHandler struct {
	artistRepo ArtistLister
}

func NewArtistHandler(repo ArtistLister) *ArtistHandler {
	return &ArtistHandler{artistRepo: repo}
}

func (h *ArtistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	f := models.ArtistFilters{
		Search:  q.Get("search"),
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}
	if genres := q["genre"]; len(genres) > 0 {
		f.Genres = genres
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

	result, err := h.artistRepo.ListForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list artists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
