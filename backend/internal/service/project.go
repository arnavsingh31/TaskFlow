package service

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/helpers"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type ProjectService struct {
	db              *sql.DB
	projectRepo     *repository.ProjectRepo
	idempotencyRepo *repository.IdempotencyRepo
	logger          *zap.Logger
}

func NewProjectService(db *sql.DB, projectRepo *repository.ProjectRepo, idempotencyRepo *repository.IdempotencyRepo, logger *zap.Logger) *ProjectService {
	return &ProjectService{
		db:              db,
		projectRepo:     projectRepo,
		idempotencyRepo: idempotencyRepo,
		logger:          logger,
	}
}

func (s *ProjectService) List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error) {
	total, err := s.projectRepo.Count(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count projects", zap.Error(err))
		return nil, 0, err
	}

	offset := (page - 1) * limit
	projects, err := s.projectRepo.List(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list projects", zap.Error(err))
		return nil, 0, err
	}
	if projects == nil {
		projects = []*model.Project{}
	}
	return projects, total, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id string) (*model.ProjectDetailResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to get project", zap.Error(err))
		return nil, err
	}

	return &model.ProjectDetailResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
		Tasks:       []*model.Task{},
	}, nil
}

func (s *ProjectService) Create(ctx context.Context, userID string, idempotencyKey *string, req *model.CreateProjectRequest) (*model.Project, error) {
	var project *model.Project

	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if idempotencyKey != nil {
			existingID, err := s.idempotencyRepo.Check(ctx, tx, *idempotencyKey, userID)
			if err == nil && existingID != "" {
				project, err = s.projectRepo.GetByIDTx(ctx, tx, existingID)
				return err
			}
		}

		var err error
		project, err = s.projectRepo.Insert(ctx, tx, req.Name, req.Description, userID)
		if err != nil {
			return err
		}

		if idempotencyKey != nil {
			if err := s.idempotencyRepo.Insert(ctx, tx, *idempotencyKey, userID, project.ID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to create project", zap.Error(err))
		return nil, err
	}

	s.logger.Info("project created", zap.String("project_id", project.ID), zap.String("user_id", userID))
	return project, nil
}

func (s *ProjectService) Update(ctx context.Context, userID, projectID string, req *model.UpdateProjectRequest) (*model.Project, error) {
	var project *model.Project

	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		p, err := s.projectRepo.GetForUpdate(ctx, tx, projectID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}

		if p.OwnerID != userID {
			return ErrForbidden
		}

		project, err = s.projectRepo.Update(ctx, tx, projectID, req)
		return err
	})

	if err != nil {
		if err != ErrNotFound && err != ErrForbidden {
			s.logger.Error("failed to update project", zap.Error(err))
		}
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, userID, projectID string) error {
	err := helpers.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		p, err := s.projectRepo.GetForUpdate(ctx, tx, projectID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}

		if p.OwnerID != userID {
			return ErrForbidden
		}

		if err := s.projectRepo.SoftDeleteTasks(ctx, tx, projectID); err != nil {
			return err
		}

		rows, err := s.projectRepo.SoftDelete(ctx, tx, projectID)
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
			s.logger.Error("failed to delete project", zap.Error(err))
		}
		return err
	}

	s.logger.Info("project deleted", zap.String("project_id", projectID))
	return nil
}

func (s *ProjectService) Stats(ctx context.Context, projectID string) (*model.ProjectStatsResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to get project for stats", zap.Error(err))
		return nil, err
	}

	statusCounts, err := s.projectRepo.GetStatsByStatus(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get status stats", zap.Error(err))
		return nil, err
	}

	assigneeCounts, err := s.projectRepo.GetStatsByAssignee(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get assignee stats", zap.Error(err))
		return nil, err
	}

	// Build response — defaults to 0 for missing statuses
	response := &model.ProjectStatsResponse{
		Todo:       0,
		InProgress: 0,
		Done:       0,
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case "todo":
			response.Todo = sc.Count
		case "in_progress":
			response.InProgress = sc.Count
		case "done":
			response.Done = sc.Count
		}
	}

	byAssignee := make([]*model.AssigneeStat, 0, len(assigneeCounts))
	for _, ac := range assigneeCounts {
		name := "Unassigned"
		if ac.Name != nil {
			name = *ac.Name
		}
		byAssignee = append(byAssignee, &model.AssigneeStat{
			AssigneeID: ac.AssigneeID,
			Name:       name,
			Count:      ac.Count,
		})
	}
	response.ByAssignee = byAssignee

	return response, nil
}
