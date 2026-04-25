package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	JWTSecret           string
	Port                string
	FrontendURL         string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("failed to load .env", "path", "../.env", "error", err)
	}

	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  os.Getenv("SPOTIFY_REDIRECT_URI"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		Port:                os.Getenv("PORT"),
		FrontendURL:         os.Getenv("FRONTEND_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.SpotifyClientID == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID is required")
	}
	if cfg.SpotifyClientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_SECRET is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "http://127.0.0.1:5173"
	}
	if cfg.SpotifyRedirectURI == "" {
		cfg.SpotifyRedirectURI = cfg.FrontendURL + "/api/auth/callback"
	}

	return cfg, nil
}
