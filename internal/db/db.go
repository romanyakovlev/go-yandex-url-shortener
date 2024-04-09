package db

import (
	"database/sql"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

func InitDB(DatabaseDSN string, sugar *logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}
	return db, nil
}
