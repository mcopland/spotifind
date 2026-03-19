package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type AlbumRepo struct {
	db *pgxpool.Pool
}

func NewAlbumRepo(db *pgxpool.Pool) *AlbumRepo {
	return &AlbumRepo{db: db}
}

func (r *AlbumRepo) Upsert(ctx context.Context, a *models.Album) (*models.Album, error) {
	query := `
		INSERT INTO albums (spotify_id, name, album_type, release_date, release_year, total_tracks, image_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			name         = EXCLUDED.name,
			album_type   = EXCLUDED.album_type,
			release_date = EXCLUDED.release_date,
			release_year = EXCLUDED.release_year,
			total_tracks = EXCLUDED.total_tracks,
			image_url    = EXCLUDED.image_url,
			updated_at   = NOW()
		RETURNING id, spotify_id, name, album_type, release_date, release_year, total_tracks, image_url, created_at, updated_at`

	row := r.db.QueryRow(ctx, query, a.SpotifyID, a.Name, a.AlbumType, a.ReleaseDate, a.ReleaseYear, a.TotalTracks, a.ImageURL)
	al := &models.Album{}
	err := row.Scan(&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return al, nil
}

func (r *AlbumRepo) LinkArtist(ctx context.Context, albumID, artistID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO album_artists (album_id, artist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		albumID, artistID)
	return err
}

func (r *AlbumRepo) LinkToUser(ctx context.Context, userID, albumID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_saved_albums (user_id, album_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, albumID)
	return err
}

func (r *AlbumRepo) GetBySpotifyID(ctx context.Context, spotifyID string) (*models.Album, error) {
	query := `SELECT id, spotify_id, name, album_type, release_date, release_year, total_tracks, image_url, created_at, updated_at FROM albums WHERE spotify_id = $1`
	al := &models.Album{}
	err := r.db.QueryRow(ctx, query, spotifyID).Scan(&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return al, nil
}

func (r *AlbumRepo) ListForUser(ctx context.Context, userID int64, f models.AlbumFilters) (*models.PaginatedResult[models.Album], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	args := []any{userID}
	where := []string{"usa.user_id = $1"}
	idx := 2

	if f.Search != "" {
		where = append(where, fmt.Sprintf("al.name ILIKE $%d", idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}
	if len(f.Genres) > 0 {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM album_artists aa2
			JOIN artists ar2 ON ar2.id = aa2.artist_id
			WHERE aa2.album_id = al.id AND ar2.genres && $%d)`, idx))
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

	whereClause := "WHERE " + strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM albums al JOIN user_saved_albums usa ON al.id = usa.album_id %s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	sortCol := "al.name"
	switch f.SortBy {
	case "release_year":
		sortCol = "al.release_year"
	case "total_tracks":
		sortCol = "al.total_tracks"
	}
	sortDir := "ASC"
	if strings.ToUpper(f.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	offset := (f.Page - 1) * f.PageSize
	listQuery := fmt.Sprintf(`
		SELECT al.id, al.spotify_id, al.name, al.album_type, al.release_date, al.release_year, al.total_tracks, al.image_url, al.created_at, al.updated_at,
		       COALESCE(array_agg(DISTINCT ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_names
		FROM albums al
		JOIN user_saved_albums usa ON al.id = usa.album_id
		LEFT JOIN album_artists aa ON al.id = aa.album_id
		LEFT JOIN artists ar ON ar.id = aa.artist_id
		%s
		GROUP BY al.id
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, idx, idx+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albums := make([]models.Album, 0)
	for rows.Next() {
		al := &models.Album{}
		var artistNames []string
		err := rows.Scan(&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt, &artistNames)
		if err != nil {
			return nil, err
		}
		for _, name := range artistNames {
			al.Artists = append(al.Artists, models.Artist{Name: name})
		}
		albums = append(albums, *al)
	}

	return &models.PaginatedResult[models.Album]{
		Items:    albums,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}
