package spotify_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestGetCurrentUser_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "u1", "display_name": "Test"})
	}
	c := newTestClient(t, handler)
	u, err := c.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "u1" {
		t.Errorf("expected id u1, got %q", u.ID)
	}
	if u.DisplayName != "Test" {
		t.Errorf("expected display_name Test, got %q", u.DisplayName)
	}
}

func TestGetCurrentUser_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	_, err := c.GetCurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetSavedTracks_SingleBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"added_at": "2024-01-01", "track": map[string]any{"id": "t1", "name": "Track 1"}},
			},
			"next": "",
		})
	}
	c := newTestClient(t, handler)
	var batches [][]spotify.SavedTrackItem
	err := c.GetSavedTracks(context.Background(), func(items []spotify.SavedTrackItem) error {
		batches = append(batches, items)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(batches))
	}
	if len(batches[0]) != 1 || batches[0][0].Track.ID != "t1" {
		t.Errorf("unexpected batch contents: %+v", batches[0])
	}
}

func TestGetSavedTracks_Pagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "0" {
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"added_at": "2024-01-01", "track": map[string]any{"id": "t1"}},
				},
				"next": "has-more",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"added_at": "2024-01-02", "track": map[string]any{"id": "t2"}},
				},
				"next": "",
			})
		}
	}
	c := newTestClient(t, handler)
	var all []spotify.SavedTrackItem
	err := c.GetSavedTracks(context.Background(), func(items []spotify.SavedTrackItem) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 items across batches, got %d", len(all))
	}
}

func TestGetSavedTracks_BatchError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"added_at": "2024-01-01", "track": map[string]any{"id": "t1"}}},
			"next":  "",
		})
	}
	c := newTestClient(t, handler)
	wantErr := errors.New("batch processing failed")
	err := c.GetSavedTracks(context.Background(), func(_ []spotify.SavedTrackItem) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected batch error to propagate, got: %v", err)
	}
}

func TestGetSavedTracks_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	err := c.GetSavedTracks(context.Background(), func(_ []spotify.SavedTrackItem) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetSavedAlbums_SingleBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"added_at": "2024-01-01", "album": map[string]any{"id": "al1", "name": "Album 1"}},
			},
			"next": "",
		})
	}
	c := newTestClient(t, handler)
	var all []spotify.SavedAlbumItem
	err := c.GetSavedAlbums(context.Background(), func(items []spotify.SavedAlbumItem) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 || all[0].Album.ID != "al1" {
		t.Errorf("unexpected items: %+v", all)
	}
}

func TestGetSavedAlbums_BatchError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"added_at": "2024-01-01", "album": map[string]any{"id": "al1"}}},
			"next":  "",
		})
	}
	c := newTestClient(t, handler)
	wantErr := errors.New("batch error")
	err := c.GetSavedAlbums(context.Background(), func(_ []spotify.SavedAlbumItem) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected batch error to propagate, got: %v", err)
	}
}

func TestGetSavedAlbums_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	err := c.GetSavedAlbums(context.Background(), func(_ []spotify.SavedAlbumItem) error { return nil })
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetFollowedArtists_SingleBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"artists": map[string]any{
				"items":   []map[string]any{{"id": "a1", "name": "Artist 1"}},
				"next":    "",
				"cursors": map[string]any{"after": ""},
			},
		})
	}
	c := newTestClient(t, handler)
	var all []spotify.SpotifyArtist
	err := c.GetFollowedArtists(context.Background(), func(items []spotify.SpotifyArtist) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 || all[0].ID != "a1" {
		t.Errorf("unexpected items: %+v", all)
	}
}

func TestGetFollowedArtists_CursorPagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		after := r.URL.Query().Get("after")
		if after == "" {
			json.NewEncoder(w).Encode(map[string]any{
				"artists": map[string]any{
					"items":   []map[string]any{{"id": "a1"}},
					"next":    "has-more",
					"cursors": map[string]any{"after": "cur1"},
				},
			})
		} else if after == "cur1" {
			json.NewEncoder(w).Encode(map[string]any{
				"artists": map[string]any{
					"items":   []map[string]any{{"id": "a2"}},
					"next":    "",
					"cursors": map[string]any{"after": ""},
				},
			})
		} else {
			http.Error(w, "unexpected after param", http.StatusBadRequest)
		}
	}
	c := newTestClient(t, handler)
	var all []spotify.SpotifyArtist
	err := c.GetFollowedArtists(context.Background(), func(items []spotify.SpotifyArtist) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 artists, got %d", len(all))
	}
}

func TestGetFollowedArtists_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	err := c.GetFollowedArtists(context.Background(), func(_ []spotify.SpotifyArtist) error { return nil })
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetFollowedArtists_BatchError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"artists": map[string]any{
				"items":   []map[string]any{{"id": "a1"}},
				"next":    "",
				"cursors": map[string]any{"after": ""},
			},
		})
	}
	c := newTestClient(t, handler)
	wantErr := errors.New("batch error")
	err := c.GetFollowedArtists(context.Background(), func(_ []spotify.SpotifyArtist) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected batch error to propagate, got: %v", err)
	}
}

func TestGetPlaylists_SingleBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "pl1", "name": "My Playlist"},
			},
			"next": "",
		})
	}
	c := newTestClient(t, handler)
	var all []spotify.SpotifyPlaylist
	err := c.GetPlaylists(context.Background(), func(items []spotify.SpotifyPlaylist) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 || all[0].ID != "pl1" {
		t.Errorf("unexpected items: %+v", all)
	}
}

func TestGetPlaylists_Pagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "0" {
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{{"id": "pl1"}},
				"next":  "has-more",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{{"id": "pl2"}},
				"next":  "",
			})
		}
	}
	c := newTestClient(t, handler)
	var all []spotify.SpotifyPlaylist
	err := c.GetPlaylists(context.Background(), func(items []spotify.SpotifyPlaylist) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 playlists, got %d", len(all))
	}
}

func TestGetPlaylists_BatchError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"id": "pl1"}},
			"next":  "",
		})
	}
	c := newTestClient(t, handler)
	wantErr := errors.New("batch error")
	err := c.GetPlaylists(context.Background(), func(_ []spotify.SpotifyPlaylist) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected batch error to propagate, got: %v", err)
	}
}

func TestGetPlaylistTracks_SingleBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playlists/pl1/tracks" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"added_at": "2024-01-01", "track": map[string]any{"id": "t1", "name": "Track 1"}},
			},
			"next": "",
		})
	}
	c := newTestClient(t, handler)
	var all []spotify.PlaylistTrackItem
	err := c.GetPlaylistTracks(context.Background(), "pl1", func(items []spotify.PlaylistTrackItem) error {
		all = append(all, items...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 item, got %d", len(all))
	}
}

func TestGetPlaylistTracks_BatchError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"added_at": "2024-01-01", "track": map[string]any{"id": "t1"}},
			},
			"next": "",
		})
	}
	c := newTestClient(t, handler)
	wantErr := errors.New("batch error")
	err := c.GetPlaylistTracks(context.Background(), "pl1", func(_ []spotify.PlaylistTrackItem) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected batch error to propagate, got: %v", err)
	}
}

func TestGetPlaylistTracks_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	err := c.GetPlaylistTracks(context.Background(), "pl1", func(_ []spotify.PlaylistTrackItem) error { return nil })
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetSavedTracks_EmptyItems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "next": ""})
	}
	c := newTestClient(t, handler)
	called := false
	err := c.GetSavedTracks(context.Background(), func(_ []spotify.SavedTrackItem) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected onBatch to never be called for empty items")
	}
}

func TestGetSavedAlbums_EmptyItems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "next": ""})
	}
	c := newTestClient(t, handler)
	called := false
	err := c.GetSavedAlbums(context.Background(), func(_ []spotify.SavedAlbumItem) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected onBatch to never be called for empty items")
	}
}

func TestGetFollowedArtists_EmptyItems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"artists": map[string]any{
				"items":   []any{},
				"next":    "",
				"cursors": map[string]any{"after": ""},
			},
		})
	}
	c := newTestClient(t, handler)
	called := false
	err := c.GetFollowedArtists(context.Background(), func(_ []spotify.SpotifyArtist) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected onBatch to never be called for empty items")
	}
}

func TestGetPlaylists_EmptyItems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "next": ""})
	}
	c := newTestClient(t, handler)
	called := false
	err := c.GetPlaylists(context.Background(), func(_ []spotify.SpotifyPlaylist) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected onBatch to never be called for empty items")
	}
}

func TestGetPlaylists_APIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}
	c := newTestClient(t, handler)
	err := c.GetPlaylists(context.Background(), func(_ []spotify.SpotifyPlaylist) error { return nil })
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetPlaylistTracks_EmptyItems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "next": ""})
	}
	c := newTestClient(t, handler)
	called := false
	err := c.GetPlaylistTracks(context.Background(), "pl1", func(_ []spotify.PlaylistTrackItem) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected onBatch to never be called for empty items")
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
