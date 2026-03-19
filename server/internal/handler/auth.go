package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
	"github.com/mcopland/spotifind/internal/spotify"
)

type AuthHandler struct {
	auth        *spotify.AuthClient
	userRepo    *repository.UserRepo
	jwtSecret   string
	frontendURL string
}

func NewAuthHandler(auth *spotify.AuthClient, userRepo *repository.UserRepo, jwtSecret, frontendURL string) *AuthHandler {
	return &AuthHandler{
		auth:        auth,
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}
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

	client := spotify.NewClient(tok.AccessToken, tok.RefreshToken, spotify.TokenExpiresAt(tok.ExpiresIn), h.auth, 0, nil)
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
