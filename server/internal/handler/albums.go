package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/spotify"
)

// AlbumLister is satisfied by repository.AlbumRepo.
type AlbumLister interface {
	ListForUser(ctx context.Context, userID int64, f models.AlbumFilters) (*models.PaginatedResult[models.Album], error)
	GetBySpotifyID(ctx context.Context, spotifyID string) (*models.Album, error)
}

// AlbumTrackChecker is satisfied by repository.TrackRepo.
type AlbumTrackChecker interface {
	LikedSpotifyIDs(ctx context.Context, userID int64, spotifyIDs []string) (map[string]bool, error)
}

// AlbumUserStore is satisfied by repository.UserRepo.
type AlbumUserStore interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	UpdateTokens(ctx context.Context, id int64, accessToken, refreshToken string, expiresAt time.Time) error
}

type AlbumHandler struct {
	albumRepo  AlbumLister
	trackRepo  AlbumTrackChecker
	userRepo   AlbumUserStore
	authClient spotify.TokenRefresher
}

func NewAlbumHandler(albumRepo AlbumLister, trackRepo AlbumTrackChecker, userRepo AlbumUserStore, authClient spotify.TokenRefresher) *AlbumHandler {
	return &AlbumHandler{
		albumRepo:  albumRepo,
		trackRepo:  trackRepo,
		userRepo:   userRepo,
		authClient: authClient,
	}
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

type albumTrack struct {
	SpotifyID   string           `json:"spotify_id"`
	Name        string           `json:"name"`
	TrackNumber int              `json:"track_number"`
	DurationMs  int              `json:"duration_ms"`
	Explicit    bool             `json:"explicit"`
	Artists     []models.Artist  `json:"artists"`
	Liked       bool             `json:"liked"`
}

type albumTracksResponse struct {
	Album  *models.Album `json:"album"`
	Tracks []albumTrack  `json:"tracks"`
}

func (h *AlbumHandler) Tracks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	albumSpotifyID := chi.URLParam(r, "id")
	if albumSpotifyID == "" {
		http.Error(w, "missing album id", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	client := spotify.NewClient(
		user.AccessToken, user.RefreshToken, user.TokenExpiresAt,
		h.authClient, userID,
		func(accessToken, refreshToken string, expiresAt time.Time) error {
			return h.userRepo.UpdateTokens(r.Context(), userID, accessToken, refreshToken, expiresAt)
		},
	)

	spotifyTracks, err := client.GetAlbumTracks(r.Context(), albumSpotifyID)
	if err != nil {
		http.Error(w, "failed to fetch album tracks from Spotify", http.StatusBadGateway)
		return
	}

	spotifyIDs := make([]string, len(spotifyTracks))
	for i, t := range spotifyTracks {
		spotifyIDs[i] = t.ID
	}

	liked, err := h.trackRepo.LikedSpotifyIDs(r.Context(), userID, spotifyIDs)
	if err != nil {
		http.Error(w, "failed to check liked tracks", http.StatusInternalServerError)
		return
	}

	tracks := make([]albumTrack, len(spotifyTracks))
	for i, t := range spotifyTracks {
		artists := make([]models.Artist, len(t.Artists))
		for j, a := range t.Artists {
			artists[j] = models.Artist{
				SpotifyID: a.ID,
				Name:      a.Name,
			}
		}
		tracks[i] = albumTrack{
			SpotifyID:   t.ID,
			Name:        t.Name,
			TrackNumber: t.TrackNumber,
			DurationMs:  t.DurationMs,
			Explicit:    t.Explicit,
			Artists:     artists,
			Liked:       liked[t.ID],
		}
	}

	albumMeta, _ := h.albumRepo.GetBySpotifyID(r.Context(), albumSpotifyID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albumTracksResponse{
		Album:  albumMeta,
		Tracks: tracks,
	})
}
