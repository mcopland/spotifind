package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type TopRepo struct {
	db *pgxpool.Pool
}

func NewTopRepo(db *pgxpool.Pool) *TopRepo {
	return &TopRepo{db: db}
}

func (r *TopRepo) DeleteTopTracksForUser(ctx context.Context, userID int64, timeRange string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_top_tracks WHERE user_id = $1 AND time_range = $2`,
		userID, timeRange)
	return err
}

func (r *TopRepo) UpsertTopTrack(ctx context.Context, userID, trackID int64, rank int, timeRange string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_top_tracks (user_id, track_id, time_range, rank) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, track_id, time_range) DO UPDATE SET rank = EXCLUDED.rank`,
		userID, trackID, timeRange, rank)
	return err
}

func (r *TopRepo) DeleteTopArtistsForUser(ctx context.Context, userID int64, timeRange string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_top_artists WHERE user_id = $1 AND time_range = $2`,
		userID, timeRange)
	return err
}

func (r *TopRepo) UpsertTopArtist(ctx context.Context, userID, artistID int64, rank int, timeRange string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_top_artists (user_id, artist_id, time_range, rank) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, artist_id, time_range) DO UPDATE SET rank = EXCLUDED.rank`,
		userID, artistID, timeRange, rank)
	return err
}

func (r *TopRepo) ListTopTracksForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopTrack], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}
	if f.TimeRange == "" {
		f.TimeRange = "short_term"
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_top_tracks WHERE user_id = $1 AND time_range = $2`,
		userID, f.TimeRange).Scan(&total); err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.spotify_id, t.name, t.album_id, t.track_number, t.duration_ms, t.explicit, t.popularity,
		       t.created_at, t.updated_at,
		       al.id, al.spotify_id, al.name, al.album_type, al.release_date, al.release_year, al.total_tracks, al.image_url, al.created_at, al.updated_at,
		       COALESCE(array_agg(DISTINCT ar.name ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_names,
		       COALESCE(array_agg(ar.spotify_id ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_spotify_ids,
		       utt.rank, utt.time_range
		FROM user_top_tracks utt
		JOIN tracks t ON t.id = utt.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		LEFT JOIN track_artists ta ON ta.track_id = t.id
		LEFT JOIN artists ar ON ar.id = ta.artist_id
		WHERE utt.user_id = $1 AND utt.time_range = $2
		GROUP BY t.id, al.id, utt.rank, utt.time_range
		ORDER BY utt.rank ASC
		LIMIT $3 OFFSET $4`,
		userID, f.TimeRange, f.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TopTrack, 0)
	for rows.Next() {
		var tt models.TopTrack
		al := &models.Album{}
		var artistNames, artistSpotifyIDs []string
		err := rows.Scan(
			&tt.ID, &tt.SpotifyID, &tt.Name, &tt.AlbumID, &tt.TrackNumber, &tt.DurationMs, &tt.Explicit, &tt.Popularity,
			&tt.CreatedAt, &tt.UpdatedAt,
			&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt,
			&artistNames, &artistSpotifyIDs,
			&tt.Rank, &tt.TimeRange,
		)
		if err != nil {
			return nil, err
		}
		tt.Album = al
		for i, name := range artistNames {
			spotifyID := ""
			if i < len(artistSpotifyIDs) {
				spotifyID = artistSpotifyIDs[i]
			}
			tt.Artists = append(tt.Artists, models.Artist{Name: name, SpotifyID: spotifyID})
		}
		items = append(items, tt)
	}

	return &models.PaginatedResult[models.TopTrack]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (r *TopRepo) ListTopArtistsForUser(ctx context.Context, userID int64, f models.TopFilters) (*models.PaginatedResult[models.TopArtist], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}
	if f.TimeRange == "" {
		f.TimeRange = "short_term"
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_top_artists WHERE user_id = $1 AND time_range = $2`,
		userID, f.TimeRange).Scan(&total); err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.spotify_id, a.name, a.image_url, a.genres, a.popularity, a.followers, a.created_at, a.updated_at,
		       uta.rank, uta.time_range
		FROM user_top_artists uta
		JOIN artists a ON a.id = uta.artist_id
		WHERE uta.user_id = $1 AND uta.time_range = $2
		ORDER BY uta.rank ASC
		LIMIT $3 OFFSET $4`,
		userID, f.TimeRange, f.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TopArtist, 0)
	for rows.Next() {
		var ta models.TopArtist
		err := rows.Scan(
			&ta.ID, &ta.SpotifyID, &ta.Name, &ta.ImageURL, &ta.Genres, &ta.Popularity, &ta.Followers,
			&ta.CreatedAt, &ta.UpdatedAt,
			&ta.Rank, &ta.TimeRange,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, ta)
	}

	return &models.PaginatedResult[models.TopArtist]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
