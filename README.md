# SpotiFind

Treat your Spotify library like a database. SpotiFind syncs your listening history, saved tracks, and playlists into a local PostgreSQL database, then exposes a filterable dashboard to explore and act on that data.

| Tech Stack             |                                                                                                           |
| ---------------------- | --------------------------------------------------------------------------------------------------------- |
| **Go**                 | Concurrent sync engine, JWT auth, HTTP API; goroutines handle long-running Spotify syncs without blocking |
| **PostgreSQL**         | Stores the synced library; GIN indexes support full-text search and array-based genre filtering           |
| **React + TypeScript** | Type-safe SPA with React Query for server state and Zustand for shared filter state                       |
| **Vite**               | Fast dev server and production bundler                                                                    |
| **Docker Compose**     | Single-command local database setup                                                                       |

## Demo

_Screenshots and deployment link coming soon._

## Features

- Sync saved tracks, albums, followed artists, and playlists from Spotify
- Filter tracks by genre, release year, popularity range, and explicit content
- Full-text search across tracks, albums, and artists
- Sort any column and paginate results
- Real-time sync progress indicator
- Recently played history and top tracks/artists by time range

## Prerequisites

- Go 1.25+
- Node.js 20+
- Docker (for the database)
- A [Spotify developer app](https://developer.spotify.com/dashboard) with redirect URI `http://localhost:8080/api/auth/callback`

## Configuration

Create a `.env` file at the repo root:

```
DATABASE_URL=postgres://spotifind:spotifind@localhost:5432/spotifind?sslmode=disable
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
JWT_SECRET=a_random_secret
```

## Development

Install coverage tools (required for `make test-ci`):

```bash
make tools
```

Start the full dev environment — starts the database, runs migrations, then both servers with hot reload:

```bash
make dev
```

The Go server runs on `:8080`. The Vite dev server runs on `:5173` and proxies `/api` to the Go server.

## Make targets

| Target                  | Description                                             |
| ----------------------- | ------------------------------------------------------- |
| `make dev`              | Start DB, migrate, and run both servers with hot reload |
| `make db-up`            | Start the dev PostgreSQL container                      |
| `make db-down`          | Stop the dev PostgreSQL container                       |
| `make migrate`          | Run database migrations                                 |
| `make build`            | Build the Go binary and Vite production bundle          |
| `make test`             | Run unit tests with coverage                            |
| `make test-cover`       | Run unit tests and print per-function coverage report   |
| `make test-integration` | Run integration tests against a live test database      |
| `make test-ci`          | Run unit + integration tests and enforce coverage gate  |
| `make tools`            | Install `gocovmerge` and `go-test-coverage`             |

## Testing

Unit tests have no external dependencies:

```bash
make test
```

Integration tests spin up a PostgreSQL container automatically:

```bash
make test-integration
```

The CI target merges both coverage profiles and enforces a 60% combined threshold:

```bash
make test-ci
```

## Building

```bash
make build
```

Outputs the Go binary to `server/bin/spotifind` and the frontend bundle to `web/dist/`. In production the binary serves the frontend as embedded static files.
