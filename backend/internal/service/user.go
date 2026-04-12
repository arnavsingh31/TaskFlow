package service

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepo
	logger   *zap.Logger
}

func NewUserService(userRepo *repository.UserRepo, logger *zap.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) SearchByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.SearchByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		s.logger.Error("failed to search user", zap.Error(err))
		return nil, err
	}
	return user, nil
}
