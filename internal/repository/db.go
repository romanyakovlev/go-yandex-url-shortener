package repository

import (
	"database/sql"
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

func (r DBURLRepository) Save(randomPath string, urlStr string) {
	return
}
