package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
	logger    *zap.Logger
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, string, error) {
	existing, err := s.userRepo.FindByEmail(ctx, *req.Email)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("failed to check existing user", zap.Error(err))
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrConflict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 12)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, "", err
	}

	user, err := s.userRepo.Insert(ctx, *req.Name, *req.Email, string(hash))
	if err != nil {
		s.logger.Error("failed to insert user", zap.Error(err))
		return nil, "", err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	s.logger.Info("user registered", zap.String("user_id", user.ID))
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.User, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, *req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrUnauthorized
		}
		s.logger.Error("failed to find user", zap.Error(err))
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(*req.Password)); err != nil {
		return nil, "", ErrUnauthorized
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	s.logger.Info("user logged in", zap.String("user_id", user.ID))
	return user, token, nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("failed to sign JWT", zap.Error(err))
		return "", err
	}
	return tokenStr, nil
}
