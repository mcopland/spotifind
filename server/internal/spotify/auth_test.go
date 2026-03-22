package spotify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/spotify"
)

func TestAuthClient_AuthorizeURL(t *testing.T) {
	a := spotify.NewAuthClient("my-client-id", "secret", "http://localhost/callback")
	u := a.AuthorizeURL("my-state")

	for _, want := range []string{"client_id=my-client-id", "redirect_uri=", "state=my-state"} {
		if !strings.Contains(u, want) {
			t.Errorf("AuthorizeURL missing %q in %q", want, u)
		}
	}
}

func TestAuthClient_ExchangeCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&spotify.TokenResponse{
			AccessToken:  "access-tok",
			RefreshToken: "refresh-tok",
			ExpiresIn:    3600,
		})
	}))
	defer srv.Close()

	a := spotify.NewTestAuthClient("id", "secret", "http://localhost/cb", srv.URL)
	tok, err := a.ExchangeCode(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if tok.AccessToken != "access-tok" {
		t.Errorf("AccessToken: want %q, got %q", "access-tok", tok.AccessToken)
	}
	if tok.RefreshToken != "refresh-tok" {
		t.Errorf("RefreshToken: want %q, got %q", "refresh-tok", tok.RefreshToken)
	}
	if tok.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn: want 3600, got %d", tok.ExpiresIn)
	}
}

func TestAuthClient_RefreshToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&spotify.TokenResponse{
			AccessToken: "new-access-tok",
			ExpiresIn:   3600,
		})
	}))
	defer srv.Close()

	a := spotify.NewTestAuthClient("id", "secret", "http://localhost/cb", srv.URL)
	tok, err := a.RefreshToken(context.Background(), "old-refresh-tok")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if tok.AccessToken != "new-access-tok" {
		t.Errorf("AccessToken: want %q, got %q", "new-access-tok", tok.AccessToken)
	}
}

func TestAuthClient_TokenExpiresAt(t *testing.T) {
	before := time.Now()
	got := spotify.TokenExpiresAt(3600)
	after := time.Now()

	want := before.Add(3600 * time.Second)
	if got.Before(want.Add(-time.Second)) || got.After(after.Add(3600*time.Second+time.Second)) {
		t.Errorf("TokenExpiresAt(%d): got %v, expected around %v", 3600, got, want)
	}
}

func TestAuthClient_requestToken_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	a := spotify.NewTestAuthClient("id", "secret", "http://localhost/cb", srv.URL)
	_, err := a.ExchangeCode(context.Background(), "bad-code")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "token request failed") {
		t.Errorf("error should mention token request failed, got: %v", err)
	}
}

func TestAuthClient_requestToken_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	a := spotify.NewTestAuthClient("id", "secret", "http://localhost/cb", srv.URL)
	_, err := a.ExchangeCode(context.Background(), "some-code")
	if err == nil {
		t.Fatal("expected JSON unmarshal error, got nil")
	}
}
