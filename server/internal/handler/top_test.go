package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

func withTopUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, id)
}

type stubTopRepo struct {
	tracks  *models.PaginatedResult[models.TopTrack]
	artists *models.PaginatedResult[models.TopArtist]
	err     error
}

func (s *stubTopRepo) ListTopTracksForUser(_ context.Context, _ int64, _ models.TopFilters) (*models.PaginatedResult[models.TopTrack], error) {
	return s.tracks, s.err
}

func (s *stubTopRepo) ListTopArtistsForUser(_ context.Context, _ int64, _ models.TopFilters) (*models.PaginatedResult[models.TopArtist], error) {
	return s.artists, s.err
}

func TestTopHandler_ListTracks_OK(t *testing.T) {
	stub := &stubTopRepo{
		tracks: &models.PaginatedResult[models.TopTrack]{
			Items: []models.TopTrack{
				{Track: models.Track{ID: 1, Name: "Top Song"}, Rank: 1, TimeRange: "short_term"},
			},
			Total: 1, Page: 1, PageSize: 50,
		},
	}

	h := handler.NewTopHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/top/tracks?time_range=short_term", nil)
	req = req.WithContext(withTopUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.ListTracks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.TopTrack]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", out.Items[0].Rank)
	}
}

func TestTopHandler_ListTracks_DefaultTimeRange(t *testing.T) {
	var capturedFilters models.TopFilters
	stub := &stubTopRepo{}
	stub.tracks = &models.PaginatedResult[models.TopTrack]{Items: []models.TopTrack{}}

	h := handler.NewTopHandler(&captureTopRepo{stub: stub, captureFilters: &capturedFilters})
	req := httptest.NewRequest(http.MethodGet, "/top/tracks", nil)
	req = req.WithContext(withTopUserID(req.Context(), 1))
	rr := httptest.NewRecorder()

	h.ListTracks(rr, req)

	if capturedFilters.TimeRange != "short_term" {
		t.Errorf("expected default time_range short_term, got %s", capturedFilters.TimeRange)
	}
}

func TestTopHandler_ListArtists_OK(t *testing.T) {
	stub := &stubTopRepo{
		artists: &models.PaginatedResult[models.TopArtist]{
			Items: []models.TopArtist{
				{Artist: models.Artist{ID: 5, Name: "Top Artist"}, Rank: 1, TimeRange: "medium_term"},
			},
			Total: 1, Page: 1, PageSize: 50,
		},
	}

	h := handler.NewTopHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/top/artists?time_range=medium_term", nil)
	req = req.WithContext(withTopUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.ListArtists(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.TopArtist]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
}

func TestTopHandler_ListTracks_Unauthorized(t *testing.T) {
	h := handler.NewTopHandler(&stubTopRepo{})
	req := httptest.NewRequest(http.MethodGet, "/top/tracks", nil)
	rr := httptest.NewRecorder()

	h.ListTracks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTopHandler_ListArtists_Unauthorized(t *testing.T) {
	h := handler.NewTopHandler(&stubTopRepo{})
	req := httptest.NewRequest(http.MethodGet, "/top/artists", nil)
	rr := httptest.NewRecorder()

	h.ListArtists(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTopHandler_ListTracks_RepoError(t *testing.T) {
	stub := &stubTopRepo{err: errors.New("db error")}
	h := handler.NewTopHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/top/tracks", nil)
	req = req.WithContext(withTopUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.ListTracks(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestTopHandler_ListArtists_RepoError(t *testing.T) {
	stub := &stubTopRepo{err: errors.New("db error")}
	h := handler.NewTopHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/top/artists", nil)
	req = req.WithContext(withTopUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.ListArtists(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// captureTopRepo captures the filters passed to list methods for assertion in tests.
type captureTopRepo struct {
	stub           *stubTopRepo
	captureFilters *models.TopFilters
}

func (c *captureTopRepo) ListTopTracksForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopTrack], error) {
	*c.captureFilters = f
	return c.stub.tracks, c.stub.err
}

func (c *captureTopRepo) ListTopArtistsForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopArtist], error) {
	*c.captureFilters = f
	return c.stub.artists, c.stub.err
}
