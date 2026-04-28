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
	ID                    int64      `json:"id"`
	SpotifyID             string     `json:"spotify_id"`
	Name                  string     `json:"name"`
	AlbumID               *int64     `json:"album_id,omitempty"`
	TrackNumber           int        `json:"track_number"`
	DurationMs            int        `json:"duration_ms"`
	Explicit              bool       `json:"explicit"`
	Popularity            int        `json:"popularity"`
	Tempo                 *float64   `json:"tempo,omitempty"`
	Key                   *int       `json:"key,omitempty"`
	Mode                  *int       `json:"mode,omitempty"`
	TimeSignature         *int       `json:"time_signature,omitempty"`
	Energy                *float64   `json:"energy,omitempty"`
	Danceability          *float64   `json:"danceability,omitempty"`
	Valence               *float64   `json:"valence,omitempty"`
	Acousticness          *float64   `json:"acousticness,omitempty"`
	Instrumentalness      *float64   `json:"instrumentalness,omitempty"`
	Liveness              *float64   `json:"liveness,omitempty"`
	Speechiness           *float64   `json:"speechiness,omitempty"`
	Loudness              *float64   `json:"loudness,omitempty"`
	AudioFeaturesSyncedAt *time.Time `json:"audio_features_synced_at,omitempty"`
	Album                 *Album     `json:"album,omitempty"`
	Artists               []Artist   `json:"artists,omitempty"`
	SavedAt               *time.Time `json:"saved_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
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
	Search                string
	Genres                []string
	YearMin               *int
	YearMax               *int
	PopularityMin         *int
	PopularityMax         *int
	DurationMin           *int
	DurationMax           *int
	Explicit              *bool
	PlaylistID            string
	SavedAtMin            *time.Time
	SavedAtMax            *time.Time
	ArtistPopularityMin   *int
	ArtistPopularityMax   *int
	ArtistFollowersMin    *int
	ArtistFollowersMax    *int
	TempoMin              *float64
	TempoMax              *float64
	EnergyMin             *float64
	EnergyMax             *float64
	DanceabilityMin       *float64
	DanceabilityMax       *float64
	ValenceMin            *float64
	ValenceMax            *float64
	AcousticnessMin       *float64
	AcousticnessMax       *float64
	InstrumentalnessMin   *float64
	InstrumentalnessMax   *float64
	LivenessMin           *float64
	LivenessMax           *float64
	SpeechinessMin        *float64
	SpeechinessMax        *float64
	LoudnessMin           *float64
	LoudnessMax           *float64
	Keys                  []int
	Mode                  *int
	TimeSignatures        []int
	Page                  int
	PageSize              int
	SortBy                string
	SortDir               string
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

type RecentlyPlayedTrack struct {
	Track
	PlayedAt time.Time `json:"played_at"`
}

type TopTrack struct {
	Track
	Rank      int    `json:"rank"`
	TimeRange string `json:"time_range"`
}

type TopArtist struct {
	Artist
	Rank      int    `json:"rank"`
	TimeRange string `json:"time_range"`
}

type RecentlyPlayedFilters struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type TopFilters struct {
	TimeRange string `json:"time_range"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}
