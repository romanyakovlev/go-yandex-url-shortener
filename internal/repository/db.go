package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type DBURLRepository struct {
	DB *sql.DB
}

func (r DBURLRepository) Find(shortURL string) (string, bool) {
	var urlRow URLRow
	row := r.DB.QueryRow("SELECT uuid, short_url, original_url FROM url_rows where short_url = $1", shortURL)
	err := row.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL)
	if err != nil {
		return "", false
	}
	return urlRow.OriginalURL, true
}

func (r DBURLRepository) Save(url models.URLToSave) error {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	_, err := r.DB.Exec(query, uuid.New(), url.RandomPath, url.URLStr)
	if err != nil {
		return err
	}
	return nil
}

func (r DBURLRepository) BatchSave(urls []models.URLToSave) error {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	for _, url := range urls {
		_, err := r.DB.Exec(query, uuid.New(), url.RandomPath, url.URLStr)
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
