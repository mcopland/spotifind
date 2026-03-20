package spotify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/spotify"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *spotify.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return spotify.NewTestClient(srv.URL, "test-token")
}

func TestGetRecentlyPlayed_HappyPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/player/recently-played" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"played_at": "2024-01-01T12:00:00Z",
					"track": map[string]any{
						"id":   "track1",
						"name": "Test Track",
					},
				},
			},
		})
	}
	c := newTestClient(t, handler)
	items, err := c.GetRecentlyPlayed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Track.ID != "track1" {
		t.Errorf("expected track id track1, got %s", items[0].Track.ID)
	}
	if items[0].PlayedAt != "2024-01-01T12:00:00Z" {
		t.Errorf("unexpected played_at: %s", items[0].PlayedAt)
	}
}

func TestGetRecentlyPlayed_Empty(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}
	c := newTestClient(t, handler)
	items, err := c.GetRecentlyPlayed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestGetRecentlyPlayed_NilTrack(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"played_at": "2024-01-01T12:00:00Z",
					"track":     nil,
				},
				{
					"played_at": "2024-01-02T12:00:00Z",
					"track": map[string]any{
						"id":   "track2",
						"name": "Valid Track",
					},
				},
			},
		})
	}
	c := newTestClient(t, handler)
	items, err := c.GetRecentlyPlayed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// nil track item should be filtered out
	if len(items) != 1 {
		t.Fatalf("expected 1 item after filtering nil track, got %d", len(items))
	}
	if items[0].Track.ID != "track2" {
		t.Errorf("expected track2, got %s", items[0].Track.ID)
	}
}

func TestGetRecentlyPlayed_EmptyTrackID(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"played_at": "2024-01-01T12:00:00Z",
					"track":     map[string]any{"id": "", "name": "No ID"},
				},
				{
					"played_at": "2024-01-02T12:00:00Z",
					"track":     map[string]any{"id": "valid", "name": "Valid"},
				},
			},
		})
	}
	c := newTestClient(t, handler)
	items, err := c.GetRecentlyPlayed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after filtering empty track id, got %d", len(items))
	}
}

func TestGetRecentlyPlayed_NonOKStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	_, err := c.GetRecentlyPlayed(context.Background())
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestGetTopTracks_HappyPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("time_range") != "short_term" {
			http.Error(w, "wrong time_range", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "t1", "name": "Top Track 1"},
				{"id": "t2", "name": "Top Track 2"},
			},
		})
	}
	c := newTestClient(t, handler)
	tracks, err := c.GetTopTracks(context.Background(), "short_term")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
	if tracks[0].ID != "t1" {
		t.Errorf("expected t1, got %s", tracks[0].ID)
	}
}

func TestGetTopTracks_Empty(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}
	c := newTestClient(t, handler)
	tracks, err := c.GetTopTracks(context.Background(), "medium_term")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(tracks))
	}
}

func TestGetTopTracks_NonOKStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}
	c := newTestClient(t, handler)
	_, err := c.GetTopTracks(context.Background(), "short_term")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestGetTopArtists_HappyPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("time_range") != "long_term" {
			http.Error(w, "wrong time_range", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "a1", "name": "Top Artist 1", "genres": []string{"rock"}},
			},
		})
	}
	c := newTestClient(t, handler)
	artists, err := c.GetTopArtists(context.Background(), "long_term")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(artists))
	}
	if artists[0].ID != "a1" {
		t.Errorf("expected a1, got %s", artists[0].ID)
	}
}

func TestGetTopArtists_NonOKStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	c := newTestClient(t, handler)
	_, err := c.GetTopArtists(context.Background(), "short_term")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

// Ensure NewTestClient accepts a zero expiry without triggering a refresh.
func TestNewTestClient_NoRefresh(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer srv.Close()

	c := spotify.NewTestClient(srv.URL, "tok")
	_, err := c.GetTopTracks(context.Background(), "short_term")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected test server to be called")
	}
	_ = time.Now() // suppress import
}
