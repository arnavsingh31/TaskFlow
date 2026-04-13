package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/taskflow/backend/internal/model"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

const taskColumnsJoin = "t.id, t.title, t.description, t.status, t.priority, t.project_id, t.assignee_id, u.name, u.email, t.created_by, t.due_date, t.deleted_at, t.created_at, t.updated_at"
const taskColumnsSimple = "id, title, description, status, priority, project_id, assignee_id, created_by, due_date, deleted_at, created_at, updated_at"

func (r *TaskRepo) ListByProject(ctx context.Context, projectID string, status *string, assigneeID *string, limit, offset int) ([]*model.Task, error) {
	query := "SELECT " + taskColumnsJoin + " FROM tasks t LEFT JOIN users u ON t.assignee_id = u.id WHERE t.project_id = $1 AND t.deleted_at IS NULL"
	args := []any{projectID}
	i := 2

	if status != nil {
		query += fmt.Sprintf(" AND t.status = $%d", i)
		args = append(args, *status)
		i++
	}
	if assigneeID != nil {
		query += fmt.Sprintf(" AND t.assignee_id = $%d", i)
		args = append(args, *assigneeID)
		i++
	}

	query += " ORDER BY t.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t, err := scanTaskJoinFromRows(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepo) CountByProject(ctx context.Context, projectID string, status *string, assigneeID *string) (int, error) {
	query := "SELECT COUNT(*) FROM tasks WHERE project_id = $1 AND deleted_at IS NULL"
	args := []any{projectID}
	i := 2

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *status)
		i++
	}
	if assigneeID != nil {
		query += fmt.Sprintf(" AND assignee_id = $%d", i)
		args = append(args, *assigneeID)
		i++
	}

	var total int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

func (r *TaskRepo) GetByID(ctx context.Context, tx *sql.Tx, id string) (*model.Task, error) {
	row := tx.QueryRowContext(ctx,
		"SELECT "+taskColumnsJoin+" FROM tasks t LEFT JOIN users u ON t.assignee_id = u.id WHERE t.id = $1 AND t.deleted_at IS NULL",
		id,
	)
	return scanTaskJoinFromRow(row)
}

func (r *TaskRepo) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*model.Task, error) {
	row := tx.QueryRowContext(ctx,
		"SELECT "+taskColumnsSimple+" FROM tasks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE",
		id,
	)
	return scanTaskSimpleFromRow(row)
}

func (r *TaskRepo) Insert(ctx context.Context, tx *sql.Tx, req *model.CreateTaskRequest, projectID, createdBy string) (*model.Task, error) {
	row := tx.QueryRowContext(ctx,
		`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date)
		 VALUES ($1, $2, COALESCE($3, 'todo'), COALESCE($4, 'medium'), $5, $6, $7, $8)
		 RETURNING `+taskColumnsSimple,
		req.Title, req.Description, req.Status, req.Priority, projectID, req.AssigneeID, createdBy, req.DueDate,
	)
	task, err := scanTaskSimpleFromRow(row)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, tx, task.ID)
}

func (r *TaskRepo) Update(ctx context.Context, tx *sql.Tx, id string, req *model.UpdateTaskRequest) (*model.Task, error) {
	sets := []string{}
	args := []any{}
	i := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", i))
		args = append(args, *req.Title)
		i++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *req.Description)
		i++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *req.Status)
		i++
	}
	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", i))
		args = append(args, *req.Priority)
		i++
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			sets = append(sets, fmt.Sprintf("assignee_id = $%d", i))
			args = append(args, nil)
		} else {
			sets = append(sets, fmt.Sprintf("assignee_id = $%d", i))
			args = append(args, *req.AssigneeID)
		}
		i++
	}
	if req.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date = $%d", i))
		args = append(args, *req.DueDate)
		i++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, tx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE tasks SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s",
		strings.Join(sets, ", "), i, taskColumnsSimple,
	)

	row := tx.QueryRowContext(ctx, query, args...)
	task, err := scanTaskSimpleFromRow(row)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, tx, task.ID)
}

func (r *TaskRepo) SoftDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := tx.ExecContext(ctx,
		"UPDATE tasks SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanTaskJoinFromRow scans a task with assignee info from a JOIN query (sql.Row)
func scanTaskJoinFromRow(row *sql.Row) (*model.Task, error) {
	t := &model.Task{}
	var assigneeID, assigneeName, assigneeEmail, description, dueDate sql.NullString
	err := row.Scan(
		&t.ID, &t.Title, &description, &t.Status, &t.Priority,
		&t.ProjectID, &assigneeID, &assigneeName, &assigneeEmail, &t.CreatedBy, &dueDate,
		&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if assigneeName.Valid {
		t.AssigneeName = &assigneeName.String
	}
	if assigneeEmail.Valid {
		t.AssigneeEmail = &assigneeEmail.String
	}
	if description.Valid {
		t.Description = &description.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.String
	}
	return t, nil
}

// scanTaskJoinFromRows scans a task with assignee info from a JOIN query (sql.Rows)
func scanTaskJoinFromRows(rows *sql.Rows) (*model.Task, error) {
	t := &model.Task{}
	var assigneeID, assigneeName, assigneeEmail, description, dueDate sql.NullString
	err := rows.Scan(
		&t.ID, &t.Title, &description, &t.Status, &t.Priority,
		&t.ProjectID, &assigneeID, &assigneeName, &assigneeEmail, &t.CreatedBy, &dueDate,
		&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if assigneeName.Valid {
		t.AssigneeName = &assigneeName.String
	}
	if assigneeEmail.Valid {
		t.AssigneeEmail = &assigneeEmail.String
	}
	if description.Valid {
		t.Description = &description.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.String
	}
	return t, nil
}

// scanTaskSimpleFromRow scans a task without assignee info (for FOR UPDATE / RETURNING)
func scanTaskSimpleFromRow(row *sql.Row) (*model.Task, error) {
	t := &model.Task{}
	var assigneeID, description, dueDate sql.NullString
	err := row.Scan(
		&t.ID, &t.Title, &description, &t.Status, &t.Priority,
		&t.ProjectID, &assigneeID, &t.CreatedBy, &dueDate,
		&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if description.Valid {
		t.Description = &description.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.String
	}
	return t, nil
}
