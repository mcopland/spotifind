package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubRefresher struct {
	tok *TokenResponse
	err error
}

func (s *stubRefresher) RefreshToken(_ context.Context, _ string) (*TokenResponse, error) {
	return s.tok, s.err
}

func TestClient_Get_RefreshesExpiredToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer srv.Close()

	stub := &stubRefresher{tok: &TokenResponse{AccessToken: "new-token", ExpiresIn: 3600}}
	c := &Client{
		apiBase:      srv.URL,
		accessToken:  "old-token",
		refreshToken: "refresh-tok",
		expiresAt:    time.Now().Add(-time.Hour),
		auth:         stub,
	}

	var out map[string]string
	if err := c.Get(context.Background(), "/test", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotAuth != "Bearer new-token" {
		t.Errorf("expected Authorization: Bearer new-token, got %q", gotAuth)
	}
}

func TestClient_Get_RefreshError(t *testing.T) {
	wantErr := errors.New("refresh failed")
	stub := &stubRefresher{err: wantErr}
	c := &Client{
		apiBase:      "http://127.0.0.1:1",
		accessToken:  "old-token",
		refreshToken: "refresh-tok",
		expiresAt:    time.Now().Add(-time.Hour),
		auth:         stub,
	}

	var out map[string]string
	err := c.Get(context.Background(), "/test", &out)
	if err == nil {
		t.Fatal("expected error when refresh fails")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error to wrap stub error, got: %v", err)
	}
}

func TestClient_Get_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := &Client{
		apiBase:     srv.URL,
		accessToken: "tok",
		expiresAt:   time.Now().Add(time.Hour),
	}

	var out map[string]string
	err := c.Get(ctx, "/test", &out)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestClient_refresh_UpdatesToken(t *testing.T) {
	stub := &stubRefresher{tok: &TokenResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		ExpiresIn:    3600,
	}}
	c := &Client{
		accessToken:  "old-access",
		refreshToken: "old-refresh",
		auth:         stub,
	}

	if err := c.refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if c.accessToken != "new-access" {
		t.Errorf("accessToken: want new-access, got %q", c.accessToken)
	}
	if c.refreshToken != "new-refresh" {
		t.Errorf("refreshToken: want new-refresh, got %q", c.refreshToken)
	}
	if c.expiresAt.IsZero() {
		t.Error("expiresAt should be updated after refresh")
	}
}

func TestClient_refresh_SkipsEmptyRefreshToken(t *testing.T) {
	stub := &stubRefresher{tok: &TokenResponse{
		AccessToken:  "new-access",
		RefreshToken: "",
		ExpiresIn:    3600,
	}}
	c := &Client{
		accessToken:  "old-access",
		refreshToken: "old-refresh",
		auth:         stub,
	}

	if err := c.refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if c.refreshToken != "old-refresh" {
		t.Errorf("refreshToken should be unchanged when new token is empty, got %q", c.refreshToken)
	}
}

func TestClient_refresh_OnRefreshCalled(t *testing.T) {
	var calledAccess, calledRefresh string
	var calledExpires time.Time

	stub := &stubRefresher{tok: &TokenResponse{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		ExpiresIn:    3600,
	}}
	c := &Client{
		accessToken:  "old-access",
		refreshToken: "old-refresh",
		auth:         stub,
		onRefresh: func(a, r string, e time.Time) error {
			calledAccess = a
			calledRefresh = r
			calledExpires = e
			return nil
		},
	}

	if err := c.refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if calledAccess != "new-access" {
		t.Errorf("onRefresh access: want new-access, got %q", calledAccess)
	}
	if calledRefresh != "new-refresh" {
		t.Errorf("onRefresh refresh: want new-refresh, got %q", calledRefresh)
	}
	if calledExpires.IsZero() {
		t.Error("onRefresh expiresAt should be non-zero")
	}
}

func TestClient_Get_429ThenSuccess(t *testing.T) {
	var callCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer srv.Close()

	c := &Client{
		apiBase:     srv.URL,
		accessToken: "tok",
		expiresAt:   time.Now().Add(time.Hour),
	}

	var out map[string]string
	if err := c.Get(context.Background(), "/test", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestClient_Get_UnmarshalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := &Client{
		apiBase:     srv.URL,
		accessToken: "tok",
		expiresAt:   time.Now().Add(time.Hour),
	}

	var out map[string]string
	err := c.Get(context.Background(), "/test", &out)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestClient_refresh_OnRefreshError(t *testing.T) {
	wantErr := errors.New("callback error")
	stub := &stubRefresher{tok: &TokenResponse{AccessToken: "new-access", ExpiresIn: 3600}}
	c := &Client{
		accessToken:  "old-access",
		refreshToken: "old-refresh",
		auth:         stub,
		onRefresh: func(_, _ string, _ time.Time) error {
			return wantErr
		},
	}

	err := c.refresh(context.Background())
	if err == nil {
		t.Fatal("expected error from onRefresh callback")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected callback error, got: %v", err)
	}
}
