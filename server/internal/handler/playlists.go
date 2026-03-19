package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/repository"
)

type PlaylistHandler struct {
	playlistRepo *repository.PlaylistRepo
}

func NewPlaylistHandler(playlistRepo *repository.PlaylistRepo) *PlaylistHandler {
	return &PlaylistHandler{playlistRepo: playlistRepo}
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
