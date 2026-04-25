//go:build integration

package sync_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	stdsync "sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
	"github.com/mcopland/spotifind/internal/spotify"
	syncsvc "github.com/mcopland/spotifind/internal/sync"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://spotifind:spotifind@localhost:5433/spotifind_test"
	}

	var err error
	testDB, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	mg, err := migrate.New("file://../database/migrations", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator for %s: %v\n", dbURL, err)
		os.Exit(1)
	}
	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
	srcErr, dbErr := mg.Close()
	if srcErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close migrate source: %v\n", srcErr)
	}
	if dbErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close migrate db: %v\n", dbErr)
	}

	if err := truncateAll(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to truncate tables: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func truncateAll(ctx context.Context) error {
	_, err := testDB.Exec(ctx, `
		TRUNCATE
			user_top_artists, user_top_tracks, user_recently_played,
			playlist_tracks, user_playlists, user_followed_artists,
			user_saved_albums, user_saved_tracks, sync_jobs,
			track_artists, album_artists,
			tracks, playlists, albums, artists, users
		CASCADE`)
	if err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}
	return nil
}

// resetDB truncates the database at the start of a test.
func resetDB(t *testing.T) {
	t.Helper()
	if err := truncateAll(context.Background()); err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// fakeSpotify is an httptest handler that impersonates the Spotify Web API and
// the Spotify accounts token endpoint. Individual paths can be forced to return
// an error status via failOn.
type fakeSpotify struct {
	mu   stdsync.Mutex
	fail map[string]int
}

func newFakeSpotify() *fakeSpotify {
	return &fakeSpotify{fail: map[string]int{}}
}

func (f *fakeSpotify) failOn(pathFragment string, status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fail[pathFragment] = status
}

func (f *fakeSpotify) forcedStatus(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	for frag, status := range f.fail {
		if strings.Contains(path, frag) {
			return status
		}
	}
	return 0
}

func (f *fakeSpotify) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if status := f.forcedStatus(r.URL.Path); status != 0 {
		http.Error(w, "forced failure", status)
		return
	}

	switch {
	case r.URL.Path == "/api/token":
		w.Header().Set("Content-Type", "application/json")
		writeBody(w, refreshTokenJSON)
	case r.URL.Path == "/v1/me":
		writeBody(w, currentUserJSON)
	case r.URL.Path == "/v1/me/tracks":
		writeBody(w, savedTracksJSON)
	case r.URL.Path == "/v1/me/albums":
		writeBody(w, savedAlbumsJSON)
	case r.URL.Path == "/v1/me/following":
		writeBody(w, followedArtistsJSON)
	case r.URL.Path == "/v1/me/playlists":
		writeBody(w, playlistsJSON)
	case strings.HasPrefix(r.URL.Path, "/v1/playlists/") && strings.HasSuffix(r.URL.Path, "/tracks"):
		writeBody(w, playlistTracksJSON)
	case r.URL.Path == "/v1/me/player/recently-played":
		writeBody(w, recentlyPlayedJSON)
	case r.URL.Path == "/v1/me/top/tracks":
		writeBody(w, topTracksJSON)
	case r.URL.Path == "/v1/me/top/artists":
		writeBody(w, topArtistsJSON)
	default:
		http.NotFound(w, r)
	}
}

func writeBody(w http.ResponseWriter, body string) {
	if _, err := w.Write([]byte(body)); err != nil {
		fmt.Fprintf(os.Stderr, "fakeSpotify write failed: %v\n", err)
	}
}

// rewriteTransport rewrites requests bound for Spotify hosts to the test server.
// This lets us exercise the real production spotify.Client (which hardcodes the
// api.spotify.com base URL) without touching production code.
type rewriteTransport struct {
	target *url.URL
	inner  http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	if host == "api.spotify.com" || host == "accounts.spotify.com" {
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = rt.target.Scheme
		cloned.URL.Host = rt.target.Host
		cloned.Host = rt.target.Host
		return rt.inner.RoundTrip(cloned)
	}
	return rt.inner.RoundTrip(req)
}

// installRewriteClient swaps http.DefaultClient for the duration of the test so
// that the production spotify.Client reaches our fake server instead of the
// real Spotify API.
func installRewriteClient(t *testing.T, serverURL string) {
	t.Helper()
	parsed, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse server url %q: %v", serverURL, err)
	}
	prev := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: &rewriteTransport{target: parsed, inner: http.DefaultTransport}}
	t.Cleanup(func() {
		http.DefaultClient = prev
	})
}

func newService() *syncsvc.Service {
	authClient := spotify.NewAuthClient("test-client-id", "test-client-secret", "http://localhost/callback")
	return syncsvc.NewService(
		repository.NewTrackRepo(testDB),
		repository.NewAlbumRepo(testDB),
		repository.NewArtistRepo(testDB),
		repository.NewPlaylistRepo(testDB),
		repository.NewSyncRepo(testDB),
		repository.NewUserRepo(testDB),
		authClient,
		repository.NewRecentlyPlayedRepo(testDB),
		repository.NewTopRepo(testDB),
	)
}

func upsertTestUser(t *testing.T, spotifyID string, tokenExpiresAt time.Time) *models.User {
	t.Helper()
	repo := repository.NewUserRepo(testDB)
	u, err := repo.Upsert(context.Background(), &models.User{
		SpotifyID:      spotifyID,
		DisplayName:    "Test User",
		Email:          spotifyID + "@example.com",
		AccessToken:    "stale-access",
		RefreshToken:   "stale-refresh",
		TokenExpiresAt: tokenExpiresAt,
	})
	if err != nil {
		t.Fatalf("failed to upsert test user %q: %v", spotifyID, err)
	}
	return u
}

func waitForSyncTerminal(t *testing.T, syncRepo *repository.SyncRepo, userID int64, timeout time.Duration) *models.SyncJob {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last *models.SyncJob
	for {
		job, err := syncRepo.GetLatestForUser(context.Background(), userID)
		if err != nil {
			t.Fatalf("failed to get latest sync job for user %d: %v", userID, err)
		}
		last = job
		if job.Status == "completed" || job.Status == "failed" {
			return job
		}
		if time.Now().After(deadline) {
			t.Fatalf("sync did not reach terminal state within %s, last status %q", timeout, last.Status)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func countRows(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := testDB.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count query %q failed: %v", query, err)
	}
	return n
}

func TestStartSync_UnknownUserReturnsError(t *testing.T) {
	resetDB(t)
	svc := newService()

	_, err := svc.StartSync(999999)
	if err == nil {
		t.Fatalf("expected error for unknown user, got nil")
	}

	// No sync_jobs row should have been created because the FK on user_id fails.
	if got := countRows(t, `SELECT COUNT(*) FROM sync_jobs`); got != 0 {
		t.Fatalf("expected 0 sync_jobs rows, got %d", got)
	}
}

func TestRunSync_HappyPath(t *testing.T) {
	resetDB(t)

	fake := newFakeSpotify()
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	installRewriteClient(t, server.URL)

	user := upsertTestUser(t, "happy-user", time.Now().Add(time.Hour))
	svc := newService()

	jobID, err := svc.StartSync(user.ID)
	if err != nil {
		t.Fatalf("StartSync: %v", err)
	}
	if jobID == 0 {
		t.Fatalf("expected non-zero job id")
	}

	syncRepo := repository.NewSyncRepo(testDB)
	job := waitForSyncTerminal(t, syncRepo, user.ID, 10*time.Second)

	if job.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q (error=%v)", job.Status, job.Error)
	}
	if job.SyncedItems <= 0 {
		t.Fatalf("expected synced_items > 0, got %d", job.SyncedItems)
	}

	// Core entities
	if got := countRows(t, `SELECT COUNT(*) FROM tracks WHERE spotify_id = 'sp-track-1'`); got != 1 {
		t.Errorf("saved track not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM tracks WHERE spotify_id = 'sp-track-2'`); got != 1 {
		t.Errorf("playlist track not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM tracks WHERE spotify_id = 'sp-track-3'`); got != 1 {
		t.Errorf("recently played track not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM tracks WHERE spotify_id = 'sp-track-4'`); got != 1 {
		t.Errorf("top track not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM albums WHERE spotify_id = 'sp-album-1'`); got != 1 {
		t.Errorf("saved album not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM albums WHERE spotify_id = 'sp-album-2'`); got != 1 {
		t.Errorf("followed album not inserted: count=%d", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM playlists WHERE spotify_id = 'sp-playlist-1'`); got != 1 {
		t.Errorf("playlist not inserted: count=%d", got)
	}

	// User-linked rows
	if got := countRows(t, `SELECT COUNT(*) FROM user_saved_tracks WHERE user_id = $1`, user.ID); got != 1 {
		t.Errorf("user_saved_tracks rows = %d, want 1", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM user_saved_albums WHERE user_id = $1`, user.ID); got != 1 {
		t.Errorf("user_saved_albums rows = %d, want 1", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM user_followed_artists WHERE user_id = $1`, user.ID); got != 1 {
		t.Errorf("user_followed_artists rows = %d, want 1", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM user_playlists WHERE user_id = $1`, user.ID); got != 1 {
		t.Errorf("user_playlists rows = %d, want 1", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM playlist_tracks`); got != 1 {
		t.Errorf("playlist_tracks rows = %d, want 1", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM user_recently_played WHERE user_id = $1`, user.ID); got != 1 {
		t.Errorf("user_recently_played rows = %d, want 1", got)
	}
	// Top tracks/artists loop over 3 time ranges.
	if got := countRows(t, `SELECT COUNT(*) FROM user_top_tracks WHERE user_id = $1`, user.ID); got != 3 {
		t.Errorf("user_top_tracks rows = %d, want 3 (one per time range)", got)
	}
	if got := countRows(t, `SELECT COUNT(*) FROM user_top_artists WHERE user_id = $1`, user.ID); got != 3 {
		t.Errorf("user_top_artists rows = %d, want 3 (one per time range)", got)
	}

	// last_synced_at should now be set on the user row.
	userRepo := repository.NewUserRepo(testDB)
	refreshed, err := userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to re-read user: %v", err)
	}
	if refreshed.LastSyncedAt == nil {
		t.Errorf("expected last_synced_at to be set after happy-path sync")
	}
}

func TestRunSync_TokenRefreshUpdatesUser(t *testing.T) {
	resetDB(t)

	fake := newFakeSpotify()
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	installRewriteClient(t, server.URL)

	// Expired token forces a refresh on the first API call.
	user := upsertTestUser(t, "refresh-user", time.Now().Add(-time.Hour))
	svc := newService()

	if _, err := svc.StartSync(user.ID); err != nil {
		t.Fatalf("StartSync: %v", err)
	}

	syncRepo := repository.NewSyncRepo(testDB)
	job := waitForSyncTerminal(t, syncRepo, user.ID, 10*time.Second)
	if job.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q (error=%v)", job.Status, job.Error)
	}

	userRepo := repository.NewUserRepo(testDB)
	refreshed, err := userRepo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to re-read user: %v", err)
	}
	if refreshed.AccessToken != "refreshed-access" {
		t.Errorf("access_token = %q, want %q", refreshed.AccessToken, "refreshed-access")
	}
	if refreshed.RefreshToken != "refreshed-refresh" {
		t.Errorf("refresh_token = %q, want %q", refreshed.RefreshToken, "refreshed-refresh")
	}
	if !refreshed.TokenExpiresAt.After(time.Now()) {
		t.Errorf("expected token_expires_at in the future, got %v", refreshed.TokenExpiresAt)
	}
}

func TestRunSync_StageFailuresNameStage(t *testing.T) {
	cases := []struct {
		name        string
		failPath    string
		wantMsgPart string
	}{
		{"tracks", "/v1/me/tracks", "sync tracks"},
		{"albums", "/v1/me/albums", "sync albums"},
		{"artists", "/v1/me/following", "sync artists"},
		{"playlists", "/v1/me/playlists", "sync playlists"},
		{"recently-played", "/v1/me/player/recently-played", "sync recently played"},
		{"top-tracks", "/v1/me/top/tracks", "sync top tracks"},
		{"top-artists", "/v1/me/top/artists", "sync top artists"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)

			fake := newFakeSpotify()
			fake.failOn(tc.failPath, http.StatusInternalServerError)
			server := httptest.NewServer(fake)
			t.Cleanup(server.Close)
			installRewriteClient(t, server.URL)

			user := upsertTestUser(t, "fail-"+tc.name, time.Now().Add(time.Hour))
			svc := newService()
			if _, err := svc.StartSync(user.ID); err != nil {
				t.Fatalf("StartSync: %v", err)
			}

			syncRepo := repository.NewSyncRepo(testDB)
			job := waitForSyncTerminal(t, syncRepo, user.ID, 10*time.Second)
			if job.Status != "failed" {
				t.Fatalf("expected status 'failed', got %q", job.Status)
			}
			if job.Error == nil {
				t.Fatalf("expected non-nil error on failed sync")
			}
			if !strings.Contains(*job.Error, tc.wantMsgPart) {
				t.Fatalf("error %q does not contain %q", *job.Error, tc.wantMsgPart)
			}
		})
	}
}

// ---- Spotify API response fixtures ----

const refreshTokenJSON = `{"access_token":"refreshed-access","refresh_token":"refreshed-refresh","expires_in":3600,"token_type":"Bearer"}`

const currentUserJSON = `{"id":"spotify-u1","display_name":"Test"}`

const savedTracksJSON = `{
  "items": [
    {
      "added_at": "2024-01-01T00:00:00Z",
      "track": {
        "id": "sp-track-1",
        "name": "Song One",
        "track_number": 1,
        "duration_ms": 180000,
        "explicit": false,
        "popularity": 50,
        "album": {
          "id": "sp-album-1",
          "name": "Album One",
          "album_type": "album",
          "release_date": "2020-05-15",
          "total_tracks": 10,
          "images": [{"url": "http://img/a1.jpg"}],
          "artists": [{"id": "sp-artist-1", "name": "Artist One", "genres": ["pop"], "popularity": 60, "followers": {"total": 1000}, "images": []}]
        },
        "artists": [{"id": "sp-artist-1", "name": "Artist One", "genres": ["pop"], "popularity": 60, "followers": {"total": 1000}, "images": []}]
      }
    }
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 50
}`

const savedAlbumsJSON = `{
  "items": [
    {
      "added_at": "2024-01-02T00:00:00Z",
      "album": {
        "id": "sp-album-2",
        "name": "Album Two",
        "album_type": "album",
        "release_date": "2019",
        "total_tracks": 5,
        "images": [{"url": "http://img/a2.jpg"}],
        "artists": [{"id": "sp-artist-2", "name": "Artist Two", "genres": ["rock"], "popularity": 70, "followers": {"total": 2000}, "images": []}]
      }
    }
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 50
}`

const followedArtistsJSON = `{
  "artists": {
    "items": [
      {"id": "sp-artist-3", "name": "Artist Three", "genres": ["jazz"], "popularity": 40, "followers": {"total": 500}, "images": []}
    ],
    "next": "",
    "cursors": {"after": ""}
  }
}`

const playlistsJSON = `{
  "items": [
    {
      "id": "sp-playlist-1",
      "name": "My Playlist",
      "description": "a test playlist",
      "public": true,
      "collaborative": false,
      "snapshot_id": "snap1",
      "images": [{"url": "http://img/pl1.jpg"}],
      "owner": {"id": "spotify-u1"},
      "tracks": {"total": 1}
    }
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 50
}`

const playlistTracksJSON = `{
  "items": [
    {
      "added_at": "2024-01-03T00:00:00Z",
      "track": {
        "id": "sp-track-2",
        "name": "Playlist Song",
        "track_number": 1,
        "duration_ms": 200000,
        "explicit": false,
        "popularity": 30,
        "album": {
          "id": "sp-album-3",
          "name": "Album Three",
          "album_type": "single",
          "release_date": "2021-03-10",
          "total_tracks": 1,
          "images": [],
          "artists": [{"id": "sp-artist-4", "name": "Artist Four", "genres": [], "popularity": 20, "followers": {"total": 100}, "images": []}]
        },
        "artists": [{"id": "sp-artist-4", "name": "Artist Four", "genres": [], "popularity": 20, "followers": {"total": 100}, "images": []}]
      }
    }
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 100
}`

const recentlyPlayedJSON = `{
  "items": [
    {
      "played_at": "2024-06-01T12:30:00Z",
      "track": {
        "id": "sp-track-3",
        "name": "Recent Song",
        "track_number": 2,
        "duration_ms": 210000,
        "explicit": true,
        "popularity": 80,
        "album": {
          "id": "sp-album-4",
          "name": "Album Four",
          "album_type": "album",
          "release_date": "2023",
          "total_tracks": 12,
          "images": [],
          "artists": [{"id": "sp-artist-5", "name": "Artist Five", "genres": ["indie"], "popularity": 55, "followers": {"total": 3000}, "images": []}]
        },
        "artists": [{"id": "sp-artist-5", "name": "Artist Five", "genres": ["indie"], "popularity": 55, "followers": {"total": 3000}, "images": []}]
      }
    }
  ]
}`

const topTracksJSON = `{
  "items": [
    {
      "id": "sp-track-4",
      "name": "Top Song",
      "track_number": 1,
      "duration_ms": 220000,
      "explicit": false,
      "popularity": 90,
      "album": {
        "id": "sp-album-5",
        "name": "Album Five",
        "album_type": "album",
        "release_date": "2022",
        "total_tracks": 8,
        "images": [],
        "artists": [{"id": "sp-artist-6", "name": "Artist Six", "genres": ["hip-hop"], "popularity": 85, "followers": {"total": 5000}, "images": []}]
      },
      "artists": [{"id": "sp-artist-6", "name": "Artist Six", "genres": ["hip-hop"], "popularity": 85, "followers": {"total": 5000}, "images": []}]
    }
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 50
}`

const topArtistsJSON = `{
  "items": [
    {"id": "sp-artist-7", "name": "Artist Seven", "genres": ["electronic"], "popularity": 75, "followers": {"total": 8000}, "images": []}
  ],
  "next": "",
  "total": 1,
  "offset": 0,
  "limit": 50
}`
