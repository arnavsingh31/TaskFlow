package main

import (
	"database/sql"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func connectDB(dbURL string, logger *zap.Logger) *sql.DB {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}

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
	return db
}

func runMigrations(db *sql.DB, logger *zap.Logger) {
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
}
