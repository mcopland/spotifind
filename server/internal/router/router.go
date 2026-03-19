package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mcopland/spotifind/internal/handler"
	"github.com/mcopland/spotifind/internal/middleware"
)

type Handlers struct {
	Auth      *handler.AuthHandler
	Tracks    *handler.TrackHandler
	Albums    *handler.AlbumHandler
	Artists   *handler.ArtistHandler
	Playlists *handler.PlaylistHandler
	Sync      *handler.SyncHandler
	Meta      *handler.MetaHandler
}

func New(h Handlers, jwtSecret, frontendURL string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", h.Auth.Login)
			r.Get("/callback", h.Auth.Callback)
			r.Post("/logout", h.Auth.Logout)
			r.With(middleware.Auth(jwtSecret)).Get("/me", h.Auth.Me)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))

			r.Get("/tracks", h.Tracks.List)
			r.Get("/albums", h.Albums.List)
			r.Get("/artists", h.Artists.List)
			r.Get("/playlists", h.Playlists.List)
			r.Get("/playlists/{id}/tracks", h.Playlists.GetTracks)

			r.Post("/sync", h.Sync.Trigger)
			r.Get("/sync/status", h.Sync.Status)

			r.Get("/genres", h.Meta.Genres)
			r.Get("/stats", h.Meta.Stats)
		})
	})

	return r
}
