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
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/spotify"
)

// stubUserFetcher implements handler.SpotifyUserFetcher for tests.
type stubUserFetcher struct {
	user *spotify.SpotifyUser
	err  error
}

func (s *stubUserFetcher) GetCurrentUser(_ context.Context) (*spotify.SpotifyUser, error) {
	return s.user, s.err
}

type stubSpotifyAuth struct {
	authorizeURL string
	token        *spotify.TokenResponse
	err          error
	lastState    string
}

func (s *stubSpotifyAuth) AuthorizeURL(state string) string {
	s.lastState = state
	return s.authorizeURL
}
func (s *stubSpotifyAuth) ExchangeCode(_ context.Context, _ string) (*spotify.TokenResponse, error) {
	return s.token, s.err
}
func (s *stubSpotifyAuth) RefreshToken(_ context.Context, _ string) (*spotify.TokenResponse, error) {
	return s.token, s.err
}

type stubUserStore struct {
	user *models.User
	err  error
}

func (s *stubUserStore) Upsert(_ context.Context, _ *models.User) (*models.User, error) {
	return s.user, s.err
}

func (s *stubUserStore) GetByID(_ context.Context, _ int64) (*models.User, error) {
	return s.user, s.err
}

func newTestAuthHandler(auth *stubSpotifyAuth, us handler.UserStore) *handler.AuthHandler {
	return handler.NewAuthHandler(auth, us, "test-secret", "http://localhost:3000/callback")
}

func loginAndGetState(t *testing.T, h *handler.AuthHandler, auth *stubSpotifyAuth) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	if auth.lastState == "" {
		t.Fatal("Login did not generate a state")
	}
	return auth.lastState
}

func TestAuthHandler_Login_Redirect(t *testing.T) {
	auth := &stubSpotifyAuth{authorizeURL: "http://accounts.spotify.com/authorize?state=x"}
	h := newTestAuthHandler(auth, &stubUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}
	if auth.lastState == "" {
		t.Error("expected state to be generated")
	}
}

func TestAuthHandler_Logout_OK(t *testing.T) {
	h := newTestAuthHandler(&stubSpotifyAuth{}, &stubUserStore{})
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			found = true
			if c.MaxAge != -1 {
				t.Errorf("session cookie MaxAge: want -1, got %d", c.MaxAge)
			}
		}
	}
	if !found {
		t.Error("session cookie not cleared")
	}
}

func TestAuthHandler_Me_OK(t *testing.T) {
	user := &models.User{ID: 42, DisplayName: "Test User", Email: "test@example.com"}
	h := newTestAuthHandler(&stubSpotifyAuth{}, &stubUserStore{user: user})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out models.User
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.DisplayName != "Test User" {
		t.Errorf("expected DisplayName Test User, got %q", out.DisplayName)
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	h := newTestAuthHandler(&stubSpotifyAuth{}, &stubUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_Me_RepoError(t *testing.T) {
	h := newTestAuthHandler(&stubSpotifyAuth{}, &stubUserStore{err: errors.New("not found")})
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(withUserID(req.Context(), 42))
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_UnknownState(t *testing.T) {
	h := newTestAuthHandler(&stubSpotifyAuth{}, &stubUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=abc&code=xyz", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_StateMismatch(t *testing.T) {
	auth := &stubSpotifyAuth{}
	h := newTestAuthHandler(auth, &stubUserStore{})
	loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=wrong&code=xyz", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_MissingCode(t *testing.T) {
	auth := &stubSpotifyAuth{}
	h := newTestAuthHandler(auth, &stubUserStore{})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state, nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_ExchangeCodeError(t *testing.T) {
	auth := &stubSpotifyAuth{err: errors.New("exchange failed")}
	h := newTestAuthHandler(auth, &stubUserStore{})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state+"&code=c", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_GetCurrentUserError(t *testing.T) {
	auth := &stubSpotifyAuth{token: &spotify.TokenResponse{AccessToken: "tok", ExpiresIn: 3600}}
	h := newTestAuthHandler(auth, &stubUserStore{})
	h.SetClientFactory(func(_ string, _ string, _ time.Time, _ handler.SpotifyAuther) handler.SpotifyUserFetcher {
		return &stubUserFetcher{err: errors.New("spotify API error")}
	})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state+"&code=c", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_UserUpsertError(t *testing.T) {
	auth := &stubSpotifyAuth{token: &spotify.TokenResponse{AccessToken: "tok", ExpiresIn: 3600}}
	h := newTestAuthHandler(auth, &stubUserStore{err: errors.New("db error")})
	h.SetClientFactory(func(_ string, _ string, _ time.Time, _ handler.SpotifyAuther) handler.SpotifyUserFetcher {
		return &stubUserFetcher{user: &spotify.SpotifyUser{ID: "sp1", DisplayName: "Test"}}
	})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state+"&code=c", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestAuthHandler_Callback_Success(t *testing.T) {
	auth := &stubSpotifyAuth{token: &spotify.TokenResponse{AccessToken: "tok", ExpiresIn: 3600}}
	h := newTestAuthHandler(auth, &stubUserStore{user: &models.User{ID: 1}})
	h.SetClientFactory(func(_ string, _ string, _ time.Time, _ handler.SpotifyAuther) handler.SpotifyUserFetcher {
		return &stubUserFetcher{user: &spotify.SpotifyUser{ID: "sp1", DisplayName: "Alice"}}
	})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state+"&code=c", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "http://localhost:3000/callback" {
		t.Errorf("expected redirect to http://localhost:3000/callback, got %q", loc)
	}
	var sessionFound bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			sessionFound = true
		}
	}
	if !sessionFound {
		t.Error("session cookie not set")
	}
}

func TestAuthHandler_Callback_Success_WithAvatar(t *testing.T) {
	auth := &stubSpotifyAuth{token: &spotify.TokenResponse{AccessToken: "tok", ExpiresIn: 3600}}
	h := newTestAuthHandler(auth, &stubUserStore{user: &models.User{ID: 1}})
	h.SetClientFactory(func(_ string, _ string, _ time.Time, _ handler.SpotifyAuther) handler.SpotifyUserFetcher {
		return &stubUserFetcher{user: &spotify.SpotifyUser{
			ID:          "sp1",
			DisplayName: "Alice",
			Images: []struct {
				URL string `json:"url"`
			}{{URL: "http://img.example.com/avatar.jpg"}},
		}}
	})
	state := loginAndGetState(t, h, auth)
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+state+"&code=c", nil)
	rr := httptest.NewRecorder()

	h.Callback(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}
	var sessionFound bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			sessionFound = true
		}
	}
	if !sessionFound {
		t.Error("session cookie not set")
	}
}
