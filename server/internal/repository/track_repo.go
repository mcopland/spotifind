package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type TrackRepo struct {
	db *pgxpool.Pool
}

func NewTrackRepo(db *pgxpool.Pool) *TrackRepo {
	return &TrackRepo{db: db}
}

func (r *TrackRepo) Upsert(ctx context.Context, t *models.Track) (*models.Track, error) {
	query := `
		INSERT INTO tracks (spotify_id, name, album_id, track_number, duration_ms, explicit, popularity, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			name         = EXCLUDED.name,
			album_id     = EXCLUDED.album_id,
			track_number = EXCLUDED.track_number,
			duration_ms  = EXCLUDED.duration_ms,
			explicit     = EXCLUDED.explicit,
			popularity   = EXCLUDED.popularity,
			updated_at   = NOW()
		RETURNING id, spotify_id, name, album_id, track_number, duration_ms, explicit, popularity, created_at, updated_at`

	tr := &models.Track{}
	err := r.db.QueryRow(ctx, query, t.SpotifyID, t.Name, t.AlbumID, t.TrackNumber, t.DurationMs, t.Explicit, t.Popularity).
		Scan(&tr.ID, &tr.SpotifyID, &tr.Name, &tr.AlbumID, &tr.TrackNumber, &tr.DurationMs, &tr.Explicit, &tr.Popularity, &tr.CreatedAt, &tr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (r *TrackRepo) LinkArtist(ctx context.Context, trackID, artistID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO track_artists (track_id, artist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		trackID, artistID)
	return err
}

func (r *TrackRepo) LinkToUser(ctx context.Context, userID, trackID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_saved_tracks (user_id, track_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, trackID)
	return err
}

func (r *TrackRepo) ListForUser(ctx context.Context, userID int64, f models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	args := []any{userID}
	where := []string{"ust.user_id = $1"}
	idx := 2

	if f.Search != "" {
		where = append(where, fmt.Sprintf("(t.name ILIKE $%d OR al.name ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}
	if len(f.Genres) > 0 {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM track_artists ta2
			JOIN artists ar2 ON ar2.id = ta2.artist_id
			WHERE ta2.track_id = t.id AND ar2.genres && $%d)`, idx))
		args = append(args, f.Genres)
		idx++
	}
	if f.YearMin != nil {
		where = append(where, fmt.Sprintf("al.release_year >= $%d", idx))
		args = append(args, *f.YearMin)
		idx++
	}
	if f.YearMax != nil {
		where = append(where, fmt.Sprintf("al.release_year <= $%d", idx))
		args = append(args, *f.YearMax)
		idx++
	}
	if f.PopularityMin != nil {
		where = append(where, fmt.Sprintf("t.popularity >= $%d", idx))
		args = append(args, *f.PopularityMin)
		idx++
	}
	if f.PopularityMax != nil {
		where = append(where, fmt.Sprintf("t.popularity <= $%d", idx))
		args = append(args, *f.PopularityMax)
		idx++
	}
	if f.DurationMin != nil {
		where = append(where, fmt.Sprintf("t.duration_ms >= $%d", idx))
		args = append(args, *f.DurationMin)
		idx++
	}
	if f.DurationMax != nil {
		where = append(where, fmt.Sprintf("t.duration_ms <= $%d", idx))
		args = append(args, *f.DurationMax)
		idx++
	}
	if f.Explicit != nil {
		where = append(where, fmt.Sprintf("t.explicit = $%d", idx))
		args = append(args, *f.Explicit)
		idx++
	}
	if f.PlaylistID != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM playlist_tracks pt
			JOIN playlists p ON p.id = pt.playlist_id
			WHERE pt.track_id = t.id AND p.spotify_id = $%d)`, idx))
		args = append(args, f.PlaylistID)
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT t.id)
		FROM tracks t
		JOIN user_saved_tracks ust ON t.id = ust.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		%s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	sortCol := "t.name"
	switch f.SortBy {
	case "popularity":
		sortCol = "t.popularity"
	case "duration":
		sortCol = "t.duration_ms"
	case "year":
		sortCol = "al.release_year"
	case "album":
		sortCol = "al.name"
	case "saved_at":
		sortCol = "ust.saved_at"
	}
	sortDir := "ASC"
	if strings.ToUpper(f.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	offset := (f.Page - 1) * f.PageSize
	listQuery := fmt.Sprintf(`
		SELECT t.id, t.spotify_id, t.name, t.album_id, t.track_number, t.duration_ms, t.explicit, t.popularity,
		       t.created_at, t.updated_at, ust.saved_at,
		       al.id, al.spotify_id, al.name, al.album_type, al.release_date, al.release_year, al.total_tracks, al.image_url, al.created_at, al.updated_at,
		       COALESCE(array_agg(DISTINCT ar.name ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_names,
		       COALESCE(array_agg(DISTINCT ar.spotify_id ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_spotify_ids
		FROM tracks t
		JOIN user_saved_tracks ust ON t.id = ust.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		LEFT JOIN track_artists ta ON ta.track_id = t.id
		LEFT JOIN artists ar ON ar.id = ta.artist_id
		%s
		GROUP BY t.id, ust.saved_at, al.id
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, idx, idx+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]models.Track, 0)
	for rows.Next() {
		tr := &models.Track{}
		al := &models.Album{}
		var artistNames, artistSpotifyIDs []string
		err := rows.Scan(
			&tr.ID, &tr.SpotifyID, &tr.Name, &tr.AlbumID, &tr.TrackNumber, &tr.DurationMs, &tr.Explicit, &tr.Popularity,
			&tr.CreatedAt, &tr.UpdatedAt, &tr.SavedAt,
			&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt,
			&artistNames, &artistSpotifyIDs,
		)
		if err != nil {
			return nil, err
		}
		tr.Album = al
		for i, name := range artistNames {
			spotifyID := ""
			if i < len(artistSpotifyIDs) {
				spotifyID = artistSpotifyIDs[i]
			}
			tr.Artists = append(tr.Artists, models.Artist{Name: name, SpotifyID: spotifyID})
		}
		tracks = append(tracks, *tr)
	}

	return &models.PaginatedResult[models.Track]{
		Items:    tracks,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
