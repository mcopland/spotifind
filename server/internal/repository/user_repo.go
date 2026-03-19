package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Upsert(ctx context.Context, u *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (spotify_id, display_name, email, avatar_url, access_token, refresh_token, token_expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			display_name     = EXCLUDED.display_name,
			email            = EXCLUDED.email,
			avatar_url       = EXCLUDED.avatar_url,
			access_token     = EXCLUDED.access_token,
			refresh_token    = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			updated_at       = NOW()
		RETURNING id, spotify_id, display_name, email, avatar_url, access_token, refresh_token, token_expires_at, last_synced_at, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		u.SpotifyID, u.DisplayName, u.Email, u.AvatarURL,
		u.AccessToken, u.RefreshToken, u.TokenExpiresAt,
	)
	return scanUser(row)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, spotify_id, display_name, email, avatar_url, access_token, refresh_token, token_expires_at, last_synced_at, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (r *UserRepo) GetBySpotifyID(ctx context.Context, spotifyID string) (*models.User, error) {
	query := `SELECT id, spotify_id, display_name, email, avatar_url, access_token, refresh_token, token_expires_at, last_synced_at, created_at, updated_at FROM users WHERE spotify_id = $1`
	row := r.db.QueryRow(ctx, query, spotifyID)
	return scanUser(row)
}

func (r *UserRepo) UpdateTokens(ctx context.Context, id int64, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `UPDATE users SET access_token = $2, refresh_token = $3, token_expires_at = $4, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, accessToken, refreshToken, expiresAt)
	return err
}

func (r *UserRepo) UpdateLastSynced(ctx context.Context, id int64) error {
	query := `UPDATE users SET last_synced_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(
		&u.ID, &u.SpotifyID, &u.DisplayName, &u.Email, &u.AvatarURL,
		&u.AccessToken, &u.RefreshToken, &u.TokenExpiresAt,
		&u.LastSyncedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
