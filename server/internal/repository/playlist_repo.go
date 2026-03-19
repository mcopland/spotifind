package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type PlaylistRepo struct {
	db *pgxpool.Pool
}

func NewPlaylistRepo(db *pgxpool.Pool) *PlaylistRepo {
	return &PlaylistRepo{db: db}
}

func (r *PlaylistRepo) Upsert(ctx context.Context, p *models.Playlist) (*models.Playlist, error) {
	query := `
		INSERT INTO playlists (spotify_id, name, description, owner_id, is_public, collaborative, snapshot_id, image_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			name          = EXCLUDED.name,
			description   = EXCLUDED.description,
			owner_id      = EXCLUDED.owner_id,
			is_public     = EXCLUDED.is_public,
			collaborative = EXCLUDED.collaborative,
			snapshot_id   = EXCLUDED.snapshot_id,
			image_url     = EXCLUDED.image_url,
			updated_at    = NOW()
		RETURNING id, spotify_id, name, description, owner_id, is_public, collaborative, snapshot_id, image_url, created_at, updated_at`

	pl := &models.Playlist{}
	err := r.db.QueryRow(ctx, query, p.SpotifyID, p.Name, p.Description, p.OwnerID, p.IsPublic, p.Collaborative, p.SnapshotID, p.ImageURL).
		Scan(&pl.ID, &pl.SpotifyID, &pl.Name, &pl.Description, &pl.OwnerID, &pl.IsPublic, &pl.Collaborative, &pl.SnapshotID, &pl.ImageURL, &pl.CreatedAt, &pl.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (r *PlaylistRepo) LinkToUser(ctx context.Context, userID, playlistID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_playlists (user_id, playlist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, playlistID)
	return err
}

func (r *PlaylistRepo) AddTrack(ctx context.Context, playlistID, trackID int64, position int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		playlistID, trackID, position)
	return err
}

func (r *PlaylistRepo) ClearTracks(ctx context.Context, playlistID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM playlist_tracks WHERE playlist_id = $1`, playlistID)
	return err
}

func (r *PlaylistRepo) ListForUser(ctx context.Context, userID int64) ([]models.Playlist, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.spotify_id, p.name, p.description, p.owner_id, p.is_public, p.collaborative, p.snapshot_id, p.image_url, p.created_at, p.updated_at,
		       COUNT(pt.track_id) AS track_count
		FROM playlists p
		JOIN user_playlists up ON p.id = up.playlist_id
		LEFT JOIN playlist_tracks pt ON p.id = pt.playlist_id
		WHERE up.user_id = $1
		GROUP BY p.id
		ORDER BY p.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []models.Playlist
	for rows.Next() {
		pl := models.Playlist{}
		if err := rows.Scan(&pl.ID, &pl.SpotifyID, &pl.Name, &pl.Description, &pl.OwnerID, &pl.IsPublic, &pl.Collaborative, &pl.SnapshotID, &pl.ImageURL, &pl.CreatedAt, &pl.UpdatedAt, &pl.TrackCount); err != nil {
			return nil, err
		}
		playlists = append(playlists, pl)
	}
	return playlists, nil
}

func (r *PlaylistRepo) GetTracksForPlaylist(ctx context.Context, userID int64, playlistSpotifyID string, f models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	f.PlaylistID = playlistSpotifyID
	return NewTrackRepo(r.db).ListForUser(ctx, userID, f)
}

func (r *PlaylistRepo) GetDistinctGenresForUser(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT unnest(ar.genres) AS genre
		FROM artists ar
		JOIN track_artists ta ON ar.id = ta.artist_id
		JOIN tracks t ON t.id = ta.track_id
		JOIN user_saved_tracks ust ON t.id = ust.track_id
		WHERE ust.user_id = $1
		ORDER BY genre`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func (r *PlaylistRepo) GetStats(ctx context.Context, userID int64) (*models.Stats, error) {
	stats := &models.Stats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM user_saved_tracks WHERE user_id = $1),
			(SELECT COUNT(*) FROM user_saved_albums WHERE user_id = $1),
			(SELECT COUNT(*) FROM user_followed_artists WHERE user_id = $1),
			(SELECT COUNT(*) FROM user_playlists WHERE user_id = $1)`,
		userID).Scan(&stats.Tracks, &stats.Albums, &stats.Artists, &stats.Playlists)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
