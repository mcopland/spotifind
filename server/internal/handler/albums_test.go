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

type stubAlbumRepo struct {
	result *models.PaginatedResult[models.Album]
	err    error
}

func (s *stubAlbumRepo) ListForUser(_ context.Context, _ int64, _ models.AlbumFilters) (*models.PaginatedResult[models.Album], error) {
	return s.result, s.err
}

func (s *stubAlbumRepo) GetBySpotifyID(_ context.Context, _ string) (*models.Album, error) {
	return nil, nil
}

func TestAlbumHandler_List_OK(t *testing.T) {
	stub := &stubAlbumRepo{
		result: &models.PaginatedResult[models.Album]{
			Items:    []models.Album{{ID: 1, Name: "Album A"}},
			Total:    1,
			Page:     1,
			PageSize: 50,
		},
	}
	h := handler.NewAlbumHandler(stub, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.Album]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
}

func TestAlbumHandler_List_Unauthorized(t *testing.T) {
	h := handler.NewAlbumHandler(&stubAlbumRepo{}, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAlbumHandler_List_RepoError(t *testing.T) {
	stub := &stubAlbumRepo{err: errors.New("db error")}
	h := handler.NewAlbumHandler(stub, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

type captureAlbumRepo struct {
	stub           *stubAlbumRepo
	captureFilters *models.AlbumFilters
}

func (c *captureAlbumRepo) ListForUser(_ context.Context, _ int64, f models.AlbumFilters) (*models.PaginatedResult[models.Album], error) {
	*c.captureFilters = f
	return c.stub.result, c.stub.err
}

func (c *captureAlbumRepo) GetBySpotifyID(_ context.Context, _ string) (*models.Album, error) {
	return nil, nil
}

func TestAlbumHandler_List_ParsesFilters(t *testing.T) {
	var captured models.AlbumFilters
	stub := &stubAlbumRepo{
		result: &models.PaginatedResult[models.Album]{Items: []models.Album{}},
	}
	h := handler.NewAlbumHandler(&captureAlbumRepo{stub: stub, captureFilters: &captured}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/albums?year_min=1990&year_max=2010&genre=jazz&page=2&page_size=20", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if captured.YearMin == nil || *captured.YearMin != 1990 {
		t.Errorf("YearMin: want 1990, got %v", captured.YearMin)
	}
	if captured.YearMax == nil || *captured.YearMax != 2010 {
		t.Errorf("YearMax: want 2010, got %v", captured.YearMax)
	}
	if len(captured.Genres) != 1 || captured.Genres[0] != "jazz" {
		t.Errorf("Genres: want [jazz], got %v", captured.Genres)
	}
	if captured.Page != 2 {
		t.Errorf("Page: want 2, got %d", captured.Page)
	}
	if captured.PageSize != 20 {
		t.Errorf("PageSize: want 20, got %d", captured.PageSize)
	}
}
