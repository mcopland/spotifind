package config

import (
	"fmt"
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
	_ = godotenv.Load("../.env")

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
		cfg.FrontendURL = "http://localhost:5173"
	}
	if cfg.SpotifyRedirectURI == "" {
		cfg.SpotifyRedirectURI = fmt.Sprintf("http://localhost:%s/api/auth/callback", cfg.Port)
	}

	return cfg, nil
}
