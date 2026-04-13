package repository

import (
	"context"
	"database/sql"

	"github.com/taskflow/backend/internal/model"
)

type ProjectRepo struct {
	db *sql.DB
}

func NewProjectRepo(db *sql.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) List(ctx context.Context, userID string, limit, offset int) ([]*model.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.deleted_at, p.created_at
		 FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id AND t.assignee_id = $1 AND t.deleted_at IS NULL
		 WHERE (p.owner_id = $1 OR t.assignee_id = $1) AND p.deleted_at IS NULL
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		p := &model.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.DeletedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Count(ctx context.Context, userID string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT p.id)
		 FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id AND t.assignee_id = $1 AND t.deleted_at IS NULL
		 WHERE (p.owner_id = $1 OR t.assignee_id = $1) AND p.deleted_at IS NULL`,
		userID,
	).Scan(&total)
	return total, err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*model.Project, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, owner_id, deleted_at, created_at FROM projects WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	return scanProject(row)
}

func (r *ProjectRepo) GetByIDTx(ctx context.Context, tx *sql.Tx, id string) (*model.Project, error) {
	row := tx.QueryRowContext(ctx,
		"SELECT id, name, description, owner_id, deleted_at, created_at FROM projects WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	return scanProject(row)
}

func (r *ProjectRepo) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*model.Project, error) {
	row := tx.QueryRowContext(ctx,
		"SELECT id, name, description, owner_id, deleted_at, created_at FROM projects WHERE id = $1 AND deleted_at IS NULL FOR UPDATE",
		id,
	)
	return scanProject(row)
}

func (r *ProjectRepo) Insert(ctx context.Context, tx *sql.Tx, name string, description *string, ownerID string) (*model.Project, error) {
	row := tx.QueryRowContext(ctx,
		`INSERT INTO projects (name, description, owner_id) VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, deleted_at, created_at`,
		name, description, ownerID,
	)
	return scanProject(row)
}

func (r *ProjectRepo) Update(ctx context.Context, tx *sql.Tx, id string, req *model.UpdateProjectRequest) (*model.Project, error) {
	row := tx.QueryRowContext(ctx,
		`UPDATE projects SET
		 name = COALESCE($1, name),
		 description = COALESCE($2, description)
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING id, name, description, owner_id, deleted_at, created_at`,
		req.Name, req.Description, id,
	)
	return scanProject(row)
}

func (r *ProjectRepo) SoftDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := tx.ExecContext(ctx,
		"UPDATE projects SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *ProjectRepo) SoftDeleteTasks(ctx context.Context, tx *sql.Tx, projectID string) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE tasks SET deleted_at = NOW(), updated_at = NOW() WHERE project_id = $1 AND deleted_at IS NULL",
		projectID,
	)
	return err
}

func scanProject(row *sql.Row) (*model.Project, error) {
	p := &model.Project{}
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.DeletedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepo) GetStatsByStatus(ctx context.Context, projectID string) ([]model.StatusCount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.status, COUNT(DISTINCT t.id) AS status_count
		 FROM projects AS p
		 INNER JOIN tasks AS t ON p.id = t.project_id
		 WHERE p.id = $1
		 AND p.deleted_at IS NULL
		 AND t.deleted_at IS NULL
		 GROUP BY t.status`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []model.StatusCount
	for rows.Next() {
		var sc model.StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	return counts, rows.Err()
}

func (r *ProjectRepo) GetStatsByAssignee(ctx context.Context, projectID string) ([]model.AssigneeCount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.assignee_id, u.name, COUNT(DISTINCT t.id) AS task_count
		 FROM projects AS p
		 INNER JOIN tasks AS t ON p.id = t.project_id
		 LEFT JOIN users AS u ON t.assignee_id = u.id
		 WHERE p.id = $1
		 AND p.deleted_at IS NULL
		 AND t.deleted_at IS NULL
		 GROUP BY t.assignee_id, u.name`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []model.AssigneeCount
	for rows.Next() {
		var ac model.AssigneeCount
		var assigneeID sql.NullString
		var name sql.NullString
		if err := rows.Scan(&assigneeID, &name, &ac.Count); err != nil {
			return nil, err
		}
		if assigneeID.Valid {
			ac.AssigneeID = &assigneeID.String
		}
		if name.Valid {
			ac.Name = &name.String
		}
		counts = append(counts, ac)
	}
	return counts, rows.Err()
}
