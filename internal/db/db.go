// Package db Пакет инициализации БД
package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

const migrationsDir = "./migrations"

func InitDB(DatabaseDSN string, sugar *logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}

	if DatabaseDSN != "" {
		if err := goose.Up(db, migrationsDir); err != nil {
			sugar.Fatalf("goose Up failed: %v\n", err)
		}
	}

	return db, nil
}
