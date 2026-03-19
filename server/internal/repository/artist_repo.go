package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type ArtistRepo struct {
	db *pgxpool.Pool
}

func NewArtistRepo(db *pgxpool.Pool) *ArtistRepo {
	return &ArtistRepo{db: db}
}

func (r *ArtistRepo) Upsert(ctx context.Context, a *models.Artist) (*models.Artist, error) {
	query := `
		INSERT INTO artists (spotify_id, name, image_url, genres, popularity, followers, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			name       = EXCLUDED.name,
			image_url  = EXCLUDED.image_url,
			genres     = EXCLUDED.genres,
			popularity = EXCLUDED.popularity,
			followers  = EXCLUDED.followers,
			updated_at = NOW()
		RETURNING id, spotify_id, name, image_url, genres, popularity, followers, created_at, updated_at`

	row := r.db.QueryRow(ctx, query, a.SpotifyID, a.Name, a.ImageURL, a.Genres, a.Popularity, a.Followers)
	return scanArtist(row)
}

func (r *ArtistRepo) LinkToUser(ctx context.Context, userID, artistID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_followed_artists (user_id, artist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, artistID)
	return err
}

func (r *ArtistRepo) ListForUser(ctx context.Context, userID int64, f models.ArtistFilters) (*models.PaginatedResult[models.Artist], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	args := []any{userID}
	where := []string{"ufa.user_id = $1"}
	idx := 2

	if f.Search != "" {
		where = append(where, fmt.Sprintf("a.name ILIKE $%d", idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}
	if len(f.Genres) > 0 {
		where = append(where, fmt.Sprintf("a.genres && $%d", idx))
		args = append(args, f.Genres)
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM artists a JOIN user_followed_artists ufa ON a.id = ufa.artist_id %s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	sortCol := "a.name"
	switch f.SortBy {
	case "popularity":
		sortCol = "a.popularity"
	case "followers":
		sortCol = "a.followers"
	}
	sortDir := "ASC"
	if strings.ToUpper(f.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	offset := (f.Page - 1) * f.PageSize
	listQuery := fmt.Sprintf(`
		SELECT a.id, a.spotify_id, a.name, a.image_url, a.genres, a.popularity, a.followers, a.created_at, a.updated_at
		FROM artists a
		JOIN user_followed_artists ufa ON a.id = ufa.artist_id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, idx, idx+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artists := make([]models.Artist, 0)
	for rows.Next() {
		a, err := scanArtist(rows)
		if err != nil {
			return nil, err
		}
		artists = append(artists, *a)
	}

	return &models.PaginatedResult[models.Artist]{
		Items:    artists,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (r *ArtistRepo) GetDistinctGenres(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT unnest(a.genres) AS genre
		FROM artists a
		JOIN user_followed_artists ufa ON a.id = ufa.artist_id
		WHERE ufa.user_id = $1
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

func scanArtist(row scanner) (*models.Artist, error) {
	a := &models.Artist{}
	err := row.Scan(&a.ID, &a.SpotifyID, &a.Name, &a.ImageURL, &a.Genres, &a.Popularity, &a.Followers, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}
