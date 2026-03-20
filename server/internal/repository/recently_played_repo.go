package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type RecentlyPlayedRepo struct {
	db *pgxpool.Pool
}

func NewRecentlyPlayedRepo(db *pgxpool.Pool) *RecentlyPlayedRepo {
	return &RecentlyPlayedRepo{db: db}
}

func (r *RecentlyPlayedRepo) Upsert(ctx context.Context, userID, trackID int64, playedAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_recently_played (user_id, track_id, played_at) VALUES ($1, $2, $3) ON CONFLICT (user_id, track_id, played_at) DO NOTHING`,
		userID, trackID, playedAt)
	return err
}

func (r *RecentlyPlayedRepo) ListForUser(ctx context.Context, userID int64, f models.RecentlyPlayedFilters) (*models.PaginatedResult[models.RecentlyPlayedTrack], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_recently_played WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.spotify_id, t.name, t.album_id, t.track_number, t.duration_ms, t.explicit, t.popularity,
		       t.created_at, t.updated_at,
		       al.id, al.spotify_id, al.name, al.album_type, al.release_date, al.release_year, al.total_tracks, al.image_url, al.created_at, al.updated_at,
		       COALESCE(array_agg(DISTINCT ar.name ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_names,
		       COALESCE(array_agg(DISTINCT ar.spotify_id ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_spotify_ids,
		       urp.played_at
		FROM user_recently_played urp
		JOIN tracks t ON t.id = urp.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		LEFT JOIN track_artists ta ON ta.track_id = t.id
		LEFT JOIN artists ar ON ar.id = ta.artist_id
		WHERE urp.user_id = $1
		GROUP BY t.id, al.id, urp.played_at
		ORDER BY urp.played_at DESC
		LIMIT $2 OFFSET $3`,
		userID, f.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.RecentlyPlayedTrack, 0)
	for rows.Next() {
		var rpt models.RecentlyPlayedTrack
		al := &models.Album{}
		var artistNames, artistSpotifyIDs []string
		err := rows.Scan(
			&rpt.ID, &rpt.SpotifyID, &rpt.Name, &rpt.AlbumID, &rpt.TrackNumber, &rpt.DurationMs, &rpt.Explicit, &rpt.Popularity,
			&rpt.CreatedAt, &rpt.UpdatedAt,
			&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt,
			&artistNames, &artistSpotifyIDs,
			&rpt.PlayedAt,
		)
		if err != nil {
			return nil, err
		}
		rpt.Album = al
		for i, name := range artistNames {
			spotifyID := ""
			if i < len(artistSpotifyIDs) {
				spotifyID = artistSpotifyIDs[i]
			}
			rpt.Artists = append(rpt.Artists, models.Artist{Name: name, SpotifyID: spotifyID})
		}
		items = append(items, rpt)
	}

	return &models.PaginatedResult[models.RecentlyPlayedTrack]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
