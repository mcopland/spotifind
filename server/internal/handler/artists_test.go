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

type stubArtistRepo struct {
	result *models.PaginatedResult[models.Artist]
	err    error
}

func (s *stubArtistRepo) ListForUser(_ context.Context, _ int64, _ models.ArtistFilters) (*models.PaginatedResult[models.Artist], error) {
	return s.result, s.err
}

func TestArtistHandler_List_OK(t *testing.T) {
	stub := &stubArtistRepo{
		result: &models.PaginatedResult[models.Artist]{
			Items:    []models.Artist{{ID: 1, Name: "Artist A"}},
			Total:    1,
			Page:     1,
			PageSize: 50,
		},
	}
	h := handler.NewArtistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.PaginatedResult[models.Artist]
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(out.Items))
	}
}

func TestArtistHandler_List_Unauthorized(t *testing.T) {
	h := handler.NewArtistHandler(&stubArtistRepo{})
	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestArtistHandler_List_RepoError(t *testing.T) {
	stub := &stubArtistRepo{err: errors.New("db error")}
	h := handler.NewArtistHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

type captureArtistRepo struct {
	stub           *stubArtistRepo
	captureFilters *models.ArtistFilters
}

func (c *captureArtistRepo) ListForUser(_ context.Context, _ int64, f models.ArtistFilters) (*models.PaginatedResult[models.Artist], error) {
	*c.captureFilters = f
	return c.stub.result, c.stub.err
}

func TestArtistHandler_List_ParsesFilters(t *testing.T) {
	var captured models.ArtistFilters
	stub := &stubArtistRepo{
		result: &models.PaginatedResult[models.Artist]{Items: []models.Artist{}},
	}
	h := handler.NewArtistHandler(&captureArtistRepo{stub: stub, captureFilters: &captured})

	req := httptest.NewRequest(http.MethodGet, "/artists?genre=metal&page=4&page_size=15", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(captured.Genres) != 1 || captured.Genres[0] != "metal" {
		t.Errorf("Genres: want [metal], got %v", captured.Genres)
	}
	if captured.Page != 4 {
		t.Errorf("Page: want 4, got %d", captured.Page)
	}
	if captured.PageSize != 15 {
		t.Errorf("PageSize: want 15, got %d", captured.PageSize)
	}
}
