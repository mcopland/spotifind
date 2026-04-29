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

type stubTrackRepo struct {
	result *models.PaginatedResult[models.Track]
	err    error
}

func (s *stubTrackRepo) ListForUser(_ context.Context, _ int64, _ models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	return s.result, s.err
}

func (s *stubTrackRepo) GetStatsForUser(_ context.Context, _ int64) (*models.TrackStats, error) {
	return &models.TrackStats{}, nil
}

func TestTrackHandler_List_OK(t *testing.T) {
	stub := &stubTrackRepo{
		result: &models.PaginatedResult[models.Track]{
			Items:    []models.Track{{ID: 1, Name: "Track A"}},
			Total:    1,
			Page:     1,
			PageSize: 50,
		},
	}
	h := handler.NewTrackHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/tracks", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.Track]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
}

func TestTrackHandler_List_Unauthorized(t *testing.T) {
	h := handler.NewTrackHandler(&stubTrackRepo{})
	req := httptest.NewRequest(http.MethodGet, "/tracks", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTrackHandler_List_RepoError(t *testing.T) {
	stub := &stubTrackRepo{err: errors.New("db error")}
	h := handler.NewTrackHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/tracks", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

type captureTrackRepo struct {
	stub           *stubTrackRepo
	captureFilters *models.TrackFilters
}

func (c *captureTrackRepo) ListForUser(_ context.Context, _ int64, f models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	*c.captureFilters = f
	return c.stub.result, c.stub.err
}

func (c *captureTrackRepo) GetStatsForUser(_ context.Context, _ int64) (*models.TrackStats, error) {
	return &models.TrackStats{}, nil
}

func TestTrackHandler_List_ParsesFilters(t *testing.T) {
	var captured models.TrackFilters
	stub := &stubTrackRepo{
		result: &models.PaginatedResult[models.Track]{Items: []models.Track{}},
	}
	h := handler.NewTrackHandler(&captureTrackRepo{stub: stub, captureFilters: &captured})

	url := "/tracks?search=hello&genre=rock&genre=pop&year_min=2000&year_max=2020" +
		"&popularity_min=10&popularity_max=90&duration_min=60000&duration_max=300000" +
		"&explicit=true&page=3&page_size=25&sort_by=name&sort_dir=asc&playlist=playlist-abc"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if captured.Search != "hello" {
		t.Errorf("Search: want hello, got %q", captured.Search)
	}
	if len(captured.Genres) != 2 || captured.Genres[0] != "rock" || captured.Genres[1] != "pop" {
		t.Errorf("Genres: want [rock pop], got %v", captured.Genres)
	}
	if captured.YearMin == nil || *captured.YearMin != 2000 {
		t.Errorf("YearMin: want 2000, got %v", captured.YearMin)
	}
	if captured.YearMax == nil || *captured.YearMax != 2020 {
		t.Errorf("YearMax: want 2020, got %v", captured.YearMax)
	}
	if captured.PopularityMin == nil || *captured.PopularityMin != 10 {
		t.Errorf("PopularityMin: want 10, got %v", captured.PopularityMin)
	}
	if captured.PopularityMax == nil || *captured.PopularityMax != 90 {
		t.Errorf("PopularityMax: want 90, got %v", captured.PopularityMax)
	}
	if captured.DurationMin == nil || *captured.DurationMin != 60000 {
		t.Errorf("DurationMin: want 60000, got %v", captured.DurationMin)
	}
	if captured.DurationMax == nil || *captured.DurationMax != 300000 {
		t.Errorf("DurationMax: want 300000, got %v", captured.DurationMax)
	}
	if captured.Explicit == nil || !*captured.Explicit {
		t.Errorf("Explicit: want true, got %v", captured.Explicit)
	}
	if captured.Page != 3 {
		t.Errorf("Page: want 3, got %d", captured.Page)
	}
	if captured.PageSize != 25 {
		t.Errorf("PageSize: want 25, got %d", captured.PageSize)
	}
	if captured.SortBy != "name" {
		t.Errorf("SortBy: want name, got %q", captured.SortBy)
	}
	if captured.SortDir != "asc" {
		t.Errorf("SortDir: want asc, got %q", captured.SortDir)
	}
	if captured.PlaylistID != "playlist-abc" {
		t.Errorf("PlaylistID: want playlist-abc, got %q", captured.PlaylistID)
	}
}
