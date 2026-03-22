package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/models"
)

type stubMetaRepo struct {
	genres []string
	stats  *models.Stats
	err    error
}

func (s *stubMetaRepo) GetDistinctGenresForUser(_ context.Context, _ int64) ([]string, error) {
	return s.genres, s.err
}

func (s *stubMetaRepo) GetStats(_ context.Context, _ int64) (*models.Stats, error) {
	return s.stats, s.err
}

func TestMetaHandler_Genres_OK(t *testing.T) {
	stub := &stubMetaRepo{genres: []string{"rock", "pop"}}
	h := handler.NewMetaHandler(nil, stub)
	req := httptest.NewRequest(http.MethodGet, "/meta/genres", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Genres(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []string
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 genres, got %d", len(out))
	}
}

func TestMetaHandler_Genres_NilResult(t *testing.T) {
	// repo returns nil (no genres) — handler must return [] not null
	stub := &stubMetaRepo{genres: nil}
	h := handler.NewMetaHandler(nil, stub)
	req := httptest.NewRequest(http.MethodGet, "/meta/genres", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Genres(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []string
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out == nil {
		t.Error("expected non-nil slice (empty array), got null")
	}
	if len(out) != 0 {
		t.Errorf("expected 0 genres, got %d", len(out))
	}
}

func TestMetaHandler_Genres_Unauthorized(t *testing.T) {
	h := handler.NewMetaHandler(nil, &stubMetaRepo{})
	req := httptest.NewRequest(http.MethodGet, "/meta/genres", nil)
	rr := httptest.NewRecorder()

	h.Genres(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMetaHandler_Genres_RepoError(t *testing.T) {
	stub := &stubMetaRepo{err: errors.New("db error")}
	h := handler.NewMetaHandler(nil, stub)
	req := httptest.NewRequest(http.MethodGet, "/meta/genres", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Genres(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestMetaHandler_Stats_OK(t *testing.T) {
	stub := &stubMetaRepo{
		stats: &models.Stats{Tracks: 100, Albums: 20, Artists: 10, Playlists: 5},
	}
	h := handler.NewMetaHandler(nil, stub)
	req := httptest.NewRequest(http.MethodGet, "/meta/stats", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.Stats
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Tracks != 100 {
		t.Errorf("expected 100 tracks, got %d", out.Tracks)
	}
}

func TestMetaHandler_Stats_Unauthorized(t *testing.T) {
	h := handler.NewMetaHandler(nil, &stubMetaRepo{})
	req := httptest.NewRequest(http.MethodGet, "/meta/stats", nil)
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMetaHandler_Stats_RepoError(t *testing.T) {
	stub := &stubMetaRepo{err: errors.New("db error")}
	h := handler.NewMetaHandler(nil, stub)
	req := httptest.NewRequest(http.MethodGet, "/meta/stats", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
