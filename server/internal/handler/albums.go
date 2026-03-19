package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

type AlbumHandler struct {
	albumRepo *repository.AlbumRepo
}

func NewAlbumHandler(albumRepo *repository.AlbumRepo) *AlbumHandler {
	return &AlbumHandler{albumRepo: albumRepo}
}

func (h *AlbumHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	f := models.AlbumFilters{
		Search:  q.Get("search"),
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
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

	result, err := h.albumRepo.ListForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list albums", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
