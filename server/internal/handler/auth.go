package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/spotify"
)

// SpotifyAuther is satisfied by *spotify.AuthClient.
type SpotifyAuther interface {
	AuthorizeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*spotify.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*spotify.TokenResponse, error)
}

// SpotifyUserFetcher is satisfied by *spotify.Client.
type SpotifyUserFetcher interface {
	GetCurrentUser(ctx context.Context) (*spotify.SpotifyUser, error)
}

// SpotifyClientFactory builds a SpotifyUserFetcher from token credentials.
type SpotifyClientFactory func(accessToken, refreshToken string, expiresAt time.Time, auth SpotifyAuther) SpotifyUserFetcher

// UserStore is satisfied by repository.UserRepo.
type UserStore interface {
	Upsert(ctx context.Context, u *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
}

type AuthHandler struct {
	auth          SpotifyAuther
	userRepo      UserStore
	jwtSecret     string
	frontendURL   string
	clientFactory SpotifyClientFactory
}

func NewAuthHandler(auth SpotifyAuther, userRepo UserStore, jwtSecret, frontendURL string) *AuthHandler {
	h := &AuthHandler{
		auth:        auth,
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}
	h.clientFactory = func(accessToken, refreshToken string, expiresAt time.Time, a SpotifyAuther) SpotifyUserFetcher {
		return spotify.NewClient(accessToken, refreshToken, expiresAt, a, 0, nil)
	}
	return h
}

// SetClientFactory overrides the factory used to build the Spotify client in Callback.
// Intended for use in tests.
func (h *AuthHandler) SetClientFactory(f SpotifyClientFactory) {
	h.clientFactory = f
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.Redirect(w, r, h.auth.AuthorizeURL(state), http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	tok, err := h.auth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	client := h.clientFactory(tok.AccessToken, tok.RefreshToken, spotify.TokenExpiresAt(tok.ExpiresIn), h.auth)
	spotifyUser, err := client.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "failed to get user profile", http.StatusInternalServerError)
		return
	}

	avatarURL := ""
	if len(spotifyUser.Images) > 0 {
		avatarURL = spotifyUser.Images[0].URL
	}

	user, err := h.userRepo.Upsert(r.Context(), &models.User{
		SpotifyID:      spotifyUser.ID,
		DisplayName:    spotifyUser.DisplayName,
		Email:          spotifyUser.Email,
		AvatarURL:      avatarURL,
		AccessToken:    tok.AccessToken,
		RefreshToken:   tok.RefreshToken,
		TokenExpiresAt: spotify.TokenExpiresAt(tok.ExpiresIn),
	})
	if err != nil {
		http.Error(w, "failed to save user", http.StatusInternalServerError)
		return
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     spotifyUser.ID,
		"user_id": user.ID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	signed, err := jwtToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    signed,
		MaxAge:   30 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	http.Redirect(w, r, h.frontendURL, http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
