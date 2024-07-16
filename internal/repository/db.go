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

// DBURLRepository представляет репозиторий для работы с URL в базе данных.
type DBURLRepository struct {
	db *sql.DB // db представляет подключение к базе данных.
}

// DBUserRepository представляет репозиторий для работы с пользователями в базе данных.
type DBUserRepository struct {
	db *sql.DB // db представляет подключение к базе данных.
}

// Find ищет URL по сокращенному адресу.
func (r DBURLRepository) Find(shortURL string) (models.URLRow, bool) {
	var urlRow models.URLRow
	row := r.db.QueryRow("SELECT uuid, short_url, original_url, is_deleted FROM url_rows WHERE short_url = $1", shortURL)
	err := row.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL, &urlRow.DeletedFlag)
	if err != nil {
		return models.URLRow{}, false
	}
	return urlRow, true
}

// FindByOriginalURL ищет сокращенный URL по оригинальному адресу.
func (r DBURLRepository) FindByOriginalURL(originalURL string) (string, bool) {
	var urlRow models.URLRow
	row := r.db.QueryRow("SELECT uuid, short_url, original_url FROM url_rows WHERE original_url = $1", originalURL)
	err := row.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL)
	if err != nil {
		return "", false
	}
	return urlRow.ShortURL, true
}

// FindByUserID ищет все URL, принадлежащие пользователю.
func (r *DBURLRepository) FindByUserID(userID uuid.UUID) ([]models.URLRow, bool) {
	var urlRows []models.URLRow

	rows, err := r.db.Query("SELECT uuid, short_url, original_url FROM url_rows WHERE user_id = $1", userID)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	for rows.Next() {
		var urlRow models.URLRow
		if err := rows.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OriginalURL); err != nil {
			return nil, false
		}
		urlRows = append(urlRows, urlRow)
	}

	if err := rows.Err(); err != nil {
		return nil, false
	}

	return urlRows, true
}

// Save сохраняет новый URL в базу данных.
func (r DBURLRepository) Save(url models.URLToSave) (uuid.UUID, error) {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	UUID := uuid.New()
	_, err := r.db.Exec(query, UUID, url.RandomPath, url.URLStr)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return uuid.UUID{}, &apperrors.OriginalURLAlreadyExists{URL: url.URLStr}
			default:
				return uuid.UUID{}, err
			}
		}
		return uuid.UUID{}, err
	}
	return UUID, nil
}

// BatchSave сохраняет несколько URL в базу данных одной транзакцией.
func (r DBURLRepository) BatchSave(urls []models.URLToSave) ([]uuid.UUID, error) {
	query := "INSERT INTO url_rows (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	var UUIDs []uuid.UUID
	tx, err := r.db.Begin()
	if err != nil {
		return UUIDs, err
	}

	var errs []error

	for _, url := range urls {
		UUID := uuid.New()
		_, err := r.db.Exec(query, UUID, url.RandomPath, url.URLStr)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				errs = append(errs, fmt.Errorf("rollback error: %v; original error: %v", rollbackErr, err))
				continue
			}
			errs = append(errs, err)
			continue
		}
		UUIDs = append(UUIDs, UUID)
	}

	if len(errs) > 0 {
		return UUIDs, errors.Join(errs...)
	}

	return UUIDs, tx.Commit()
}

// BatchDelete помечает URL как удаленные для указанного пользователя.
func (r *DBURLRepository) BatchDelete(urls []string, userID uuid.UUID) error {
	query := `UPDATE url_rows SET is_deleted = true WHERE user_id = $1 AND short_url = ANY($2)`

	result, err := r.db.Exec(query, userID, urls)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser обновляет пользователя для указанного URL.
func (r DBUserRepository) UpdateUser(savedURLUUID uuid.UUID, userID uuid.UUID) error {
	query := "UPDATE url_rows SET user_id = $1 WHERE uuid = $2"
	result, err := r.db.Exec(query, userID, savedURLUUID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows were updated")
	}
	return nil
}

// UpdateBatchUser обновляет пользователя для нескольких URL.
func (r *DBUserRepository) UpdateBatchUser(savedURLUUIDs []uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE url_rows SET user_id = $1 WHERE uuid = ANY($2)`

	result, err := r.db.Exec(query, userID, savedURLUUIDs)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if int(rowsAffected) != len(savedURLUUIDs) {
		return fmt.Errorf("expected to update %d rows, but %d rows were updated", len(savedURLUUIDs), rowsAffected)
	}

	return nil
}

// NewDBURLRepository создает новый экземпляр репозитория URL.
func NewDBURLRepository(db *sql.DB) (*DBURLRepository, error) {
	return &DBURLRepository{db: db}, nil
}

// NewDBUserRepository создает новый экземпляр репозитория пользователей.
func NewDBUserRepository(db *sql.DB) (*DBUserRepository, error) {
	return &DBUserRepository{db: db}, nil
}
