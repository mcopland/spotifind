package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/models"
)

type stubSyncService struct {
	jobID int64
	err   error
}

func (s *stubSyncService) StartSync(_ int64) (int64, error) {
	return s.jobID, s.err
}

type stubSyncRepo struct {
	job *models.SyncJob
	err error
}

func (s *stubSyncRepo) GetLatestForUser(_ context.Context, _ int64) (*models.SyncJob, error) {
	return s.job, s.err
}

func TestSyncHandler_Trigger_OK(t *testing.T) {
	svc := &stubSyncService{jobID: 7}
	h := handler.NewSyncHandler(svc, &stubSyncRepo{})
	req := httptest.NewRequest(http.MethodPost, "/sync/trigger", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Trigger(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
	var out map[string]int64
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out["job_id"] != 7 {
		t.Errorf("expected job_id 7, got %d", out["job_id"])
	}
}

func TestSyncHandler_Trigger_Unauthorized(t *testing.T) {
	h := handler.NewSyncHandler(&stubSyncService{}, &stubSyncRepo{})
	req := httptest.NewRequest(http.MethodPost, "/sync/trigger", nil)
	rr := httptest.NewRecorder()

	h.Trigger(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestSyncHandler_Trigger_ServiceError(t *testing.T) {
	svc := &stubSyncService{err: errors.New("service error")}
	h := handler.NewSyncHandler(svc, &stubSyncRepo{})
	req := httptest.NewRequest(http.MethodPost, "/sync/trigger", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Trigger(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestSyncHandler_Status_OK(t *testing.T) {
	job := &models.SyncJob{ID: 3, UserID: 42, Status: "completed"}
	h := handler.NewSyncHandler(&stubSyncService{}, &stubSyncRepo{job: job})
	req := httptest.NewRequest(http.MethodGet, "/sync/status", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Status(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.SyncJob
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Status != "completed" {
		t.Errorf("expected status completed, got %q", out.Status)
	}
}

func TestSyncHandler_Status_Unauthorized(t *testing.T) {
	h := handler.NewSyncHandler(&stubSyncService{}, &stubSyncRepo{})
	req := httptest.NewRequest(http.MethodGet, "/sync/status", nil)
	rr := httptest.NewRecorder()

	h.Status(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestSyncHandler_Status_NoRows(t *testing.T) {
	h := handler.NewSyncHandler(&stubSyncService{}, &stubSyncRepo{err: pgx.ErrNoRows})
	req := httptest.NewRequest(http.MethodGet, "/sync/status", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Status(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out["status"] != "none" {
		t.Errorf("expected status none, got %q", out["status"])
	}
}

func TestSyncHandler_Status_RepoError(t *testing.T) {
	h := handler.NewSyncHandler(&stubSyncService{}, &stubSyncRepo{err: errors.New("db error")})
	req := httptest.NewRequest(http.MethodGet, "/sync/status", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Status(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
