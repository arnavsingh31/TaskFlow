package service

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/helpers"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type TaskService struct {
	db              *sql.DB
	taskRepo        *repository.TaskRepo
	projectRepo     *repository.ProjectRepo
	idempotencyRepo *repository.IdempotencyRepo
	logger          *zap.Logger
}

func NewTaskService(db *sql.DB, taskRepo *repository.TaskRepo, projectRepo *repository.ProjectRepo, idempotencyRepo *repository.IdempotencyRepo, logger *zap.Logger) *TaskService {
	return &TaskService{
		db:              db,
		taskRepo:        taskRepo,
		projectRepo:     projectRepo,
		idempotencyRepo: idempotencyRepo,
		logger:          logger,
	}
}

func (s *TaskService) ListByProject(ctx context.Context, projectID string, status *string, assigneeID *string, page, limit int) ([]*model.Task, int, error) {
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, ErrNotFound
		}
		s.logger.Error("failed to verify project", zap.Error(err))
		return nil, 0, err
	}

	total, err := s.taskRepo.CountByProject(ctx, projectID, status, assigneeID)
	if err != nil {
		s.logger.Error("failed to count tasks", zap.Error(err))
		return nil, 0, err
	}

	offset := (page - 1) * limit
	tasks, err := s.taskRepo.ListByProject(ctx, projectID, status, assigneeID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list tasks", zap.Error(err))
		return nil, 0, err
	}
	if tasks == nil {
		tasks = []*model.Task{}
	}
	return tasks, total, nil
}

func (s *TaskService) Create(ctx context.Context, userID, projectID string, idempotencyKey *string, req *model.CreateTaskRequest) (*model.Task, error) {
	var task *model.Task

	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if idempotencyKey != nil {
			existingID, err := s.idempotencyRepo.Check(ctx, tx, *idempotencyKey, userID)
			if err == nil && existingID != "" {
				task, err = s.taskRepo.GetByID(ctx, tx, existingID)
				return err
			}
		}

		_, err := s.projectRepo.GetByIDTx(ctx, tx, projectID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}

		task, err = s.taskRepo.Insert(ctx, tx, req, projectID, userID)
		if err != nil {
			return err
		}

		if idempotencyKey != nil {
			if err := s.idempotencyRepo.Insert(ctx, tx, *idempotencyKey, userID, task.ID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if err != ErrNotFound {
			s.logger.Error("failed to create task", zap.Error(err))
		}
		return nil, err
	}

	s.logger.Info("task created", zap.String("task_id", task.ID), zap.String("project_id", projectID))
	return task, nil
}

func (s *TaskService) Update(ctx context.Context, userID, taskID string, req *model.UpdateTaskRequest) (*model.Task, error) {
	var task *model.Task

	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		_, err := s.taskRepo.GetForUpdate(ctx, tx, taskID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}

		task, err = s.taskRepo.Update(ctx, tx, taskID, req)
		return err
	})

	if err != nil {
		if err != ErrNotFound {
			s.logger.Error("failed to update task", zap.Error(err))
		}
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		t, err := s.taskRepo.GetForUpdate(ctx, tx, taskID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}

		project, err := s.projectRepo.GetByIDTx(ctx, tx, t.ProjectID)
		if err != nil {
			return err
		}

		isOwner := project.OwnerID == userID
		isCreator := t.CreatedBy == userID
		if !isOwner && !isCreator {
			return ErrForbidden
		}

		rows, err := s.taskRepo.SoftDelete(ctx, tx, taskID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrNotFound
		}
		return nil
	})

	if err != nil {
		if err != ErrNotFound && err != ErrForbidden {
			s.logger.Error("failed to delete task", zap.Error(err))
		}
		return err
	}

	s.logger.Info("task deleted", zap.String("task_id", taskID))
	return nil
}
