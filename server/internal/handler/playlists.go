package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// PlaylistQuerier is satisfied by repository.PlaylistRepo.
type PlaylistQuerier interface {
	ListForUser(ctx context.Context, userID int64) ([]models.Playlist, error)
	GetTracksForPlaylist(ctx context.Context, userID int64, playlistSpotifyID string, f models.TrackFilters) (*models.PaginatedResult[models.Track], error)
}

type PlaylistHandler struct {
	playlistRepo PlaylistQuerier
}

func NewPlaylistHandler(repo PlaylistQuerier) *PlaylistHandler {
	return &PlaylistHandler{playlistRepo: repo}
}

func (h *PlaylistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	playlists, err := h.playlistRepo.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}

func (h *PlaylistHandler) GetTracks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	playlistID := chi.URLParam(r, "id")
	f := parseTrackFilters(r)
	f.PlaylistID = playlistID

	result, err := h.playlistRepo.GetTracksForPlaylist(r.Context(), userID, playlistID, f)
	if err != nil {
		http.Error(w, "failed to get playlist tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
