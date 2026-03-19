package models

import "time"

type User struct {
	ID             int64      `json:"id"`
	SpotifyID      string     `json:"spotify_id"`
	DisplayName    string     `json:"display_name"`
	Email          string     `json:"email"`
	AvatarURL      string     `json:"avatar_url"`
	AccessToken    string     `json:"-"`
	RefreshToken   string     `json:"-"`
	TokenExpiresAt time.Time  `json:"-"`
	LastSyncedAt   *time.Time `json:"last_synced_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Artist struct {
	ID         int64     `json:"id"`
	SpotifyID  string    `json:"spotify_id"`
	Name       string    `json:"name"`
	ImageURL   string    `json:"image_url"`
	Genres     []string  `json:"genres"`
	Popularity int       `json:"popularity"`
	Followers  int       `json:"followers"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Album struct {
	ID          int64     `json:"id"`
	SpotifyID   string    `json:"spotify_id"`
	Name        string    `json:"name"`
	AlbumType   string    `json:"album_type"`
	ReleaseDate string    `json:"release_date"`
	ReleaseYear int       `json:"release_year"`
	TotalTracks int       `json:"total_tracks"`
	ImageURL    string    `json:"image_url"`
	Artists     []Artist  `json:"artists,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Track struct {
	ID          int64      `json:"id"`
	SpotifyID   string     `json:"spotify_id"`
	Name        string     `json:"name"`
	AlbumID     *int64     `json:"album_id,omitempty"`
	TrackNumber int        `json:"track_number"`
	DurationMs  int        `json:"duration_ms"`
	Explicit    bool       `json:"explicit"`
	Popularity  int        `json:"popularity"`
	Album       *Album     `json:"album,omitempty"`
	Artists     []Artist   `json:"artists,omitempty"`
	SavedAt     *time.Time `json:"saved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Playlist struct {
	ID            int64     `json:"id"`
	SpotifyID     string    `json:"spotify_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	OwnerID       string    `json:"owner_id"`
	IsPublic      bool      `json:"is_public"`
	Collaborative bool      `json:"collaborative"`
	SnapshotID    string    `json:"snapshot_id"`
	ImageURL      string    `json:"image_url"`
	TrackCount    int       `json:"track_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SyncJob struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Status      string     `json:"status"`
	EntityType  string     `json:"entity_type"`
	TotalItems  int        `json:"total_items"`
	SyncedItems int        `json:"synced_items"`
	Error       *string    `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TrackFilters struct {
	Search        string
	Genres        []string
	YearMin       *int
	YearMax       *int
	PopularityMin *int
	PopularityMax *int
	DurationMin   *int
	DurationMax   *int
	Explicit      *bool
	PlaylistID    string
	Page          int
	PageSize      int
	SortBy        string
	SortDir       string
}

type AlbumFilters struct {
	Search   string
	Genres   []string
	YearMin  *int
	YearMax  *int
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type ArtistFilters struct {
	Search   string
	Genres   []string
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type PaginatedResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type Stats struct {
	Tracks    int `json:"tracks"`
	Albums    int `json:"albums"`
	Artists   int `json:"artists"`
	Playlists int `json:"playlists"`
}
