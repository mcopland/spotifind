package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/models"
)

type stubPlaylistRepo struct {
	playlists []models.Playlist
	tracks    *models.PaginatedResult[models.Track]
	err       error
}

func (s *stubPlaylistRepo) ListForUser(_ context.Context, _ int64) ([]models.Playlist, error) {
	return s.playlists, s.err
}

func (s *stubPlaylistRepo) GetTracksForPlaylist(_ context.Context, _ int64, _ string, _ models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	return s.tracks, s.err
}

func TestPlaylistHandler_List_OK(t *testing.T) {
	stub := &stubPlaylistRepo{
		playlists: []models.Playlist{{ID: 1, Name: "My Playlist"}},
	}
	h := handler.NewPlaylistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/playlists", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []models.Playlist
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("expected 1 playlist, got %d", len(out))
	}
}

func TestPlaylistHandler_List_Unauthorized(t *testing.T) {
	h := handler.NewPlaylistHandler(&stubPlaylistRepo{})
	req := httptest.NewRequest(http.MethodGet, "/playlists", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestPlaylistHandler_List_RepoError(t *testing.T) {
	stub := &stubPlaylistRepo{err: errors.New("db error")}
	h := handler.NewPlaylistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/playlists", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestPlaylistHandler_GetTracks_OK(t *testing.T) {
	stub := &stubPlaylistRepo{
		tracks: &models.PaginatedResult[models.Track]{
			Items:    []models.Track{{ID: 1, Name: "Track A"}},
			Total:    1,
			Page:     1,
			PageSize: 50,
		},
	}
	h := handler.NewPlaylistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/playlists/playlist-spotify-id/tracks", nil)
	req = req.WithContext(withUserID(req.Context(), 42))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "playlist-spotify-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetTracks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.Track]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 track, got %d", len(out.Items))
	}
}

func TestPlaylistHandler_GetTracks_Unauthorized(t *testing.T) {
	h := handler.NewPlaylistHandler(&stubPlaylistRepo{})
	req := httptest.NewRequest(http.MethodGet, "/playlists/x/tracks", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "x")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetTracks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestPlaylistHandler_GetTracks_RepoError(t *testing.T) {
	stub := &stubPlaylistRepo{err: errors.New("db error")}
	h := handler.NewPlaylistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/playlists/x/tracks", nil)
	req = req.WithContext(withUserID(req.Context(), 42))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "x")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetTracks(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
