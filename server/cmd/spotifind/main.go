package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcopland/spotifind/internal/config"
	"github.com/mcopland/spotifind/internal/database"
	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/repository"
	"github.com/mcopland/spotifind/internal/router"
	"github.com/mcopland/spotifind/internal/spotify"
	syncpkg "github.com/mcopland/spotifind/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	trackRepo := repository.NewTrackRepo(pool)
	albumRepo := repository.NewAlbumRepo(pool)
	artistRepo := repository.NewArtistRepo(pool)
	playlistRepo := repository.NewPlaylistRepo(pool)
	syncRepo := repository.NewSyncRepo(pool)
	recentlyPlayedRepo := repository.NewRecentlyPlayedRepo(pool)
	topRepo := repository.NewTopRepo(pool)

	authClient := spotify.NewAuthClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRedirectURI)
	syncService := syncpkg.NewService(trackRepo, albumRepo, artistRepo, playlistRepo, syncRepo, userRepo, authClient, recentlyPlayedRepo, topRepo)

	handlers := router.Handlers{
		Auth:           handler.NewAuthHandler(authClient, userRepo, cfg.JWTSecret, cfg.FrontendURL),
		Tracks:         handler.NewTrackHandler(trackRepo),
		Albums:         handler.NewAlbumHandler(albumRepo),
		Artists:        handler.NewArtistHandler(artistRepo),
		Playlists:      handler.NewPlaylistHandler(playlistRepo),
		Sync:           handler.NewSyncHandler(syncService, syncRepo),
		Meta:           handler.NewMetaHandler(artistRepo, playlistRepo),
		RecentlyPlayed: handler.NewRecentlyPlayedHandler(recentlyPlayedRepo),
		Top:            handler.NewTopHandler(topRepo),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router.New(handlers, cfg.JWTSecret, cfg.FrontendURL),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
