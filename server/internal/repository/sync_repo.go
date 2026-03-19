package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type SyncRepo struct {
	db *pgxpool.Pool
}

func NewSyncRepo(db *pgxpool.Pool) *SyncRepo {
	return &SyncRepo{db: db}
}

func (r *SyncRepo) Create(ctx context.Context, userID int64) (*models.SyncJob, error) {
	job := &models.SyncJob{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO sync_jobs (user_id, status, entity_type)
		VALUES ($1, 'pending', 'all')
		RETURNING id, user_id, status, entity_type, total_items, synced_items, error, started_at, finished_at, created_at`,
		userID).Scan(&job.ID, &job.UserID, &job.Status, &job.EntityType, &job.TotalItems, &job.SyncedItems, &job.Error, &job.StartedAt, &job.FinishedAt, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (r *SyncRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE sync_jobs SET status = $2, started_at = CASE WHEN $2 = 'running' THEN NOW() ELSE started_at END WHERE id = $1`, id, status)
	return err
}

func (r *SyncRepo) UpdateProgress(ctx context.Context, id int64, total, synced int) error {
	_, err := r.db.Exec(ctx, `UPDATE sync_jobs SET total_items = $2, synced_items = $3 WHERE id = $1`, id, total, synced)
	return err
}

func (r *SyncRepo) Finish(ctx context.Context, id int64, status string, errMsg *string) error {
	_, err := r.db.Exec(ctx, `UPDATE sync_jobs SET status = $2, error = $3, finished_at = NOW() WHERE id = $1`, id, status, errMsg)
	return err
}

func (r *SyncRepo) GetLatestForUser(ctx context.Context, userID int64) (*models.SyncJob, error) {
	job := &models.SyncJob{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, entity_type, total_items, synced_items, error, started_at, finished_at, created_at
		FROM sync_jobs WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&job.ID, &job.UserID, &job.Status, &job.EntityType, &job.TotalItems, &job.SyncedItems, &job.Error, &job.StartedAt, &job.FinishedAt, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	return job, nil
}
