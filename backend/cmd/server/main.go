package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("failed to create logger: ", err)
	}
	defer logger.Sync()

	db, err := sql.Open("pgx", cfg.DBURL())
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer db.Close()

	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		logger.Info("waiting for database...")
		time.Sleep(1 * time.Second)
	}
	if err := db.Ping(); err != nil {
		logger.Fatal("database not ready", zap.Error(err))
	}
	logger.Info("database connected")

	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		logger.Fatal("failed to create migration driver", zap.Error(err))
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		logger.Fatal("failed to create migrator", zap.Error(err))
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("migrations applied")

	userRepo := repository.NewUserRepo(db)
	projectRepo := repository.NewProjectRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	idempotencyRepo := repository.NewIdempotencyRepo(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, logger)
	userService := service.NewUserService(userRepo, logger)
	projectService := service.NewProjectService(db, projectRepo, idempotencyRepo, logger)
	taskService := service.NewTaskService(db, taskRepo, projectRepo, idempotencyRepo, logger)

	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	projectHandler := handler.NewProjectHandler(projectService, logger)
	taskHandler := handler.NewTaskHandler(taskService, logger)

	authMW := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(cfg.JWTSecret, next)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("GET /projects", authMW(projectHandler.List))
	mux.HandleFunc("POST /projects", authMW(projectHandler.Create))
	mux.HandleFunc("GET /projects/{id}", authMW(projectHandler.Get))
	mux.HandleFunc("PATCH /projects/{id}", authMW(projectHandler.Update))
	mux.HandleFunc("DELETE /projects/{id}", authMW(projectHandler.Delete))

	mux.HandleFunc("GET /users/search", authMW(userHandler.Search))

	mux.HandleFunc("GET /projects/{id}/tasks", authMW(taskHandler.List))
	mux.HandleFunc("POST /projects/{id}/tasks", authMW(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", authMW(taskHandler.Update))
	mux.HandleFunc("DELETE /tasks/{id}", authMW(taskHandler.Delete))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"not found"}`)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORS(mux),
	}

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

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
