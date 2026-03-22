package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// TrackLister is satisfied by repository.TrackRepo.
type TrackLister interface {
	ListForUser(ctx context.Context, userID int64, f models.TrackFilters) (*models.PaginatedResult[models.Track], error)
}

type TrackHandler struct {
	trackRepo TrackLister
}

func NewTrackHandler(repo TrackLister) *TrackHandler {
	return &TrackHandler{trackRepo: repo}
}

func (h *TrackHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	f := parseTrackFilters(r)
	result, err := h.trackRepo.ListForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func parseTrackFilters(r *http.Request) models.TrackFilters {
	q := r.URL.Query()
	f := models.TrackFilters{
		Search:     q.Get("search"),
		PlaylistID: q.Get("playlist"),
		SortBy:     q.Get("sort_by"),
		SortDir:    q.Get("sort_dir"),
	}

	if genres := q["genre"]; len(genres) > 0 {
		f.Genres = genres
	}

	if v := q.Get("year_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.YearMin = &n
		}
	}
	if v := q.Get("year_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.YearMax = &n
		}
	}
	if v := q.Get("popularity_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.PopularityMin = &n
		}
	}
	if v := q.Get("popularity_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.PopularityMax = &n
		}
	}
	if v := q.Get("duration_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.DurationMin = &n
		}
	}
	if v := q.Get("duration_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.DurationMax = &n
		}
	}
	if v := q.Get("explicit"); v != "" {
		b := v == "true"
		f.Explicit = &b
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
