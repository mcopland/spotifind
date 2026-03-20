# SpotiFind

Treat your Spotify library like a database. SpotiFind syncs your listening history, saved tracks, and playlists into a local PostgreSQL database, then exposes a filterable dashboard to explore and act on that data.

| Tech Stack |  |
| --- | --- |
| **Go** | Concurrent sync engine, JWT auth, HTTP API; goroutines handle long-running Spotify syncs without blocking |
| **PostgreSQL** | Stores the synced library; GIN indexes support full-text search and array-based genre filtering |
| **React + TypeScript** | Type-safe SPA with React Query for server state and Zustand for shared filter state |
| **Vite** | Fast dev server and production bundler |
| **Docker Compose** | Single-command local database setup |

## Demo

_Screenshots and deployment link coming soon._

## Features

- Sync saved tracks, albums, followed artists, and playlists from Spotify
- Filter tracks by genre, release year, popularity range, and explicit content
- Full-text search across tracks, albums, and artists
- Sort any column and paginate results
- Real-time sync progress indicator
