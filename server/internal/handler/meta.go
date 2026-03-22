package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

// MetaQuerier is satisfied by repository.PlaylistRepo.
type MetaQuerier interface {
	GetDistinctGenresForUser(ctx context.Context, userID int64) ([]string, error)
	GetStats(ctx context.Context, userID int64) (*models.Stats, error)
}

type MetaHandler struct {
	artistRepo   *repository.ArtistRepo
	playlistRepo MetaQuerier
}

func NewMetaHandler(artistRepo *repository.ArtistRepo, playlistRepo MetaQuerier) *MetaHandler {
	return &MetaHandler{artistRepo: artistRepo, playlistRepo: playlistRepo}
}

func (h *MetaHandler) Genres(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	genres, err := h.playlistRepo.GetDistinctGenresForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get genres", http.StatusInternalServerError)
		return
	}
	if genres == nil {
		genres = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *MetaHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.playlistRepo.GetStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
