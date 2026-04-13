package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func main() {
	cfg := config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("failed to create logger: ", err)
	}
	defer logger.Sync()

	// Database
	db := connectDB(cfg.DBURL(), logger)
	defer db.Close()
	runMigrations(db, logger)

	// Repositories
	userRepo := repository.NewUserRepo(db)
	projectRepo := repository.NewProjectRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	idempotencyRepo := repository.NewIdempotencyRepo(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger)
	userService := service.NewUserService(userRepo, logger)
	projectService := service.NewProjectService(db, projectRepo, taskRepo, idempotencyRepo, logger)
	taskService := service.NewTaskService(db, taskRepo, projectRepo, idempotencyRepo, logger)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	projectHandler := handler.NewProjectHandler(projectService, logger)
	taskHandler := handler.NewTaskHandler(taskService, logger)

	// Router
	router := newRouter(cfg.JWTSecret, authHandler, userHandler, projectHandler, taskHandler)

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server stopped")
}
