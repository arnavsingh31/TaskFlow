package repository

import (
	"context"
	"database/sql"

	"github.com/taskflow/backend/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name, email, password, deleted_at, created_at FROM users WHERE email = $1 AND deleted_at IS NULL",
		email,
	)
	return scanUser(row)
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name, email, password, deleted_at, created_at FROM users WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	return scanUser(row)
}

func (r *UserRepo) Insert(ctx context.Context, name, email, hashedPassword string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)
		 RETURNING id, name, email, password, deleted_at, created_at`,
		name, email, hashedPassword,
	)
	return scanUser(row)
}

func (r *UserRepo) SearchByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.FindByEmail(ctx, email)
}

func scanUser(row *sql.Row) (*model.User, error) {
	u := &model.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.DeletedAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
