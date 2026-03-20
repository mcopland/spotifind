package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

func withUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, id)
}

// stubRecentlyPlayedRepo implements handler.RecentlyPlayedLister.
type stubRecentlyPlayedRepo struct {
	result *models.PaginatedResult[models.RecentlyPlayedTrack]
	err    error
}

func (s *stubRecentlyPlayedRepo) ListForUser(_ context.Context, _ int64, _ models.RecentlyPlayedFilters) (*models.PaginatedResult[models.RecentlyPlayedTrack], error) {
	return s.result, s.err
}

func TestRecentlyPlayedHandler_List_OK(t *testing.T) {
	stub := &stubRecentlyPlayedRepo{
		result: &models.PaginatedResult[models.RecentlyPlayedTrack]{
			Items: []models.RecentlyPlayedTrack{
				{
					Track:    models.Track{ID: 1, Name: "Song A"},
					PlayedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			Total:    1,
			Page:     1,
			PageSize: 50,
		},
	}

	h := handler.NewRecentlyPlayedHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/recently-played", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.RecentlyPlayedTrack]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
}

func TestRecentlyPlayedHandler_List_Unauthorized(t *testing.T) {
	h := handler.NewRecentlyPlayedHandler(&stubRecentlyPlayedRepo{})
	req := httptest.NewRequest(http.MethodGet, "/recently-played", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRecentlyPlayedHandler_List_RepoError(t *testing.T) {
	stub := &stubRecentlyPlayedRepo{err: errors.New("db error")}
	h := handler.NewRecentlyPlayedHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/recently-played", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// captureRecentlyPlayedRepo records the filters passed to ListForUser.
type captureRecentlyPlayedRepo struct {
	stub           *stubRecentlyPlayedRepo
	captureFilters *models.RecentlyPlayedFilters
}

func (c *captureRecentlyPlayedRepo) ListForUser(ctx context.Context, userID int64, f models.RecentlyPlayedFilters) (*models.PaginatedResult[models.RecentlyPlayedTrack], error) {
	*c.captureFilters = f
	return c.stub.result, c.stub.err
}

func TestRecentlyPlayedHandler_List_ParsesPageParams(t *testing.T) {
	var capturedFilters models.RecentlyPlayedFilters
	stub := &stubRecentlyPlayedRepo{
		result: &models.PaginatedResult[models.RecentlyPlayedTrack]{Items: []models.RecentlyPlayedTrack{}},
	}
	h := handler.NewRecentlyPlayedHandler(&captureRecentlyPlayedRepo{stub: stub, captureFilters: &capturedFilters})
	req := httptest.NewRequest(http.MethodGet, "/recently-played?page=2&page_size=10", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedFilters.Page != 2 {
		t.Errorf("expected page 2, got %d", capturedFilters.Page)
	}
	if capturedFilters.PageSize != 10 {
		t.Errorf("expected page_size 10, got %d", capturedFilters.PageSize)
	}
}
