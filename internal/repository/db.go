package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/apperrors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type DBURLRepository struct {
	db *sql.DB
}

func (r DBURLRepository) Find(shortURL string) (string, bool) {
	var urlRow URLRow
	row := r.db.QueryRow("SELECT uuid, short_url, original_url FROM url_rows where short_url = $1", shortURL)
	err := row.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL)
	if err != nil {
		return "", false
	}
	return urlRow.OriginalURL, true
}

func (r DBURLRepository) FindByOriginalURL(originalURL string) (string, bool) {
	var urlRow URLRow
	row := r.db.QueryRow("SELECT uuid, short_url, original_url FROM url_rows where original_url = $1", originalURL)
	err := row.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL)
	if err != nil {
		return "", false
	}
	return urlRow.ShortURL, true
}

func (r DBURLRepository) Save(url models.URLToSave) error {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, uuid.New(), url.RandomPath, url.URLStr)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return &apperrors.OriginalURLAlreadyExists{URL: url.URLStr}
			default:
				return err
			}
		}
		return err
	}
	return nil
}

func (r DBURLRepository) BatchSave(urls []models.URLToSave) error {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	for _, url := range urls {
		_, err := r.db.Exec(query, uuid.New(), url.RandomPath, url.URLStr)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("rollback error: %v; original error: %v", rollbackErr, err)
			}
			return err
		}
	}
	return tx.Commit()
}

func NewDBURLRepository(db *sql.DB) (*DBURLRepository, error) {
	return &DBURLRepository{db: db}, nil
}
