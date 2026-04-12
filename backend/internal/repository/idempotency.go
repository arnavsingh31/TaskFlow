package repository

import (
	"context"
	"database/sql"
)

type IdempotencyRepo struct {
	db *sql.DB
}

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

func (r *IdempotencyRepo) Check(ctx context.Context, tx *sql.Tx, key, userID string) (string, error) {
	var resourceID string
	err := tx.QueryRowContext(ctx,
		"SELECT resource_id FROM idempotency_keys WHERE key = $1 AND user_id = $2",
		key, userID,
	).Scan(&resourceID)
	return resourceID, err
}

func (r *IdempotencyRepo) Insert(ctx context.Context, tx *sql.Tx, key, userID, resourceID string) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO idempotency_keys (key, user_id, resource_id) VALUES ($1, $2, $3)",
		key, userID, resourceID,
	)
	return err
}
