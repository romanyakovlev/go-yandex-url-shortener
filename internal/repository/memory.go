package repository

import (
	"errors"

	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type MemoryURLRepository struct {
	SharedURLRows *models.SharedURLRows
}

type MemoryUserRepository struct {
	SharedURLRows *models.SharedURLRows
}

func (r *MemoryURLRepository) Save(url models.URLToSave) (uuid.UUID, error) {
	UUID := uuid.New()
	newURLRow := models.URLRow{
		UUID:        UUID,
		ShortURL:    url.RandomPath,
		OriginalURL: url.URLStr,
		DeletedFlag: false,
	}

	r.SharedURLRows.Mu.Lock()
	r.SharedURLRows.URLRows = append(r.SharedURLRows.URLRows, newURLRow)
	r.SharedURLRows.Mu.Unlock()

	return UUID, nil
}

func (r *MemoryURLRepository) BatchSave(urls []models.URLToSave) ([]uuid.UUID, error) {
	var UUIDs []uuid.UUID

	r.SharedURLRows.Mu.Lock()
	for _, url := range urls {
		UUID := uuid.New()
		newURLRow := models.URLRow{
			UUID:        UUID,
			ShortURL:    url.RandomPath,
			OriginalURL: url.URLStr,
		}
		UUIDs = append(UUIDs, UUID)
		r.SharedURLRows.URLRows = append(r.SharedURLRows.URLRows, newURLRow)
	}
	r.SharedURLRows.Mu.Unlock()

	return UUIDs, nil
}

func (r *MemoryURLRepository) Find(shortURL string) (models.URLRow, bool) {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	for _, urlRow := range r.SharedURLRows.URLRows {
		if urlRow.ShortURL == shortURL {
			return urlRow, true
		}
	}
	return models.URLRow{}, false
}

func (r *MemoryURLRepository) FindByOriginalURL(originalURL string) (string, bool) {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	for _, urlRow := range r.SharedURLRows.URLRows {
		if urlRow.OriginalURL == originalURL {
			return urlRow.ShortURL, true
		}
	}
	return "", false
}

func (r *MemoryURLRepository) FindByUserID(userID uuid.UUID) ([]models.URLRow, bool) {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	var matchedURLs []models.URLRow
	for _, urlRow := range r.SharedURLRows.URLRows {
		if urlRow.UserID == userID {
			matchedURLs = append(matchedURLs, urlRow)
		}
	}
	return matchedURLs, len(matchedURLs) > 0
}

func (r *MemoryURLRepository) BatchDelete(urls []string, userID uuid.UUID) error {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	uuidMap := make(map[string]bool)
	for _, shortURL := range urls {
		uuidMap[shortURL] = true
	}

	for i, urlRow := range r.SharedURLRows.URLRows {
		if _, exists := uuidMap[urlRow.ShortURL]; exists && urlRow.UserID == userID {
			r.SharedURLRows.URLRows[i].DeletedFlag = true
		}
	}

	return nil
}

func (r *MemoryUserRepository) UpdateUser(SavedURLUUID uuid.UUID, userID uuid.UUID) error {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	for i, urlRow := range r.SharedURLRows.URLRows {
		if urlRow.UUID == SavedURLUUID {
			r.SharedURLRows.URLRows[i].UserID = userID
			return nil
		}
	}
	return errors.New("URL not found")
}

func (r *MemoryUserRepository) UpdateBatchUser(SavedURLUUIDs []uuid.UUID, userID uuid.UUID) error {
	r.SharedURLRows.Mu.Lock()
	defer r.SharedURLRows.Mu.Unlock()

	uuidMap := make(map[uuid.UUID]bool)
	for _, id := range SavedURLUUIDs {
		uuidMap[id] = true
	}

	updated := false
	for i, urlRow := range r.SharedURLRows.URLRows {
		if _, exists := uuidMap[urlRow.UUID]; exists {
			r.SharedURLRows.URLRows[i].UserID = userID
			updated = true
		}
	}

	if !updated {
		return errors.New("no URLs updated")
	}
	return nil
}

func NewMemoryURLRepository(sharedURLRows *models.SharedURLRows) (*MemoryURLRepository, error) {
	return &MemoryURLRepository{SharedURLRows: sharedURLRows}, nil
}

func NewMemoryUserRepository(sharedURLRows *models.SharedURLRows) (*MemoryUserRepository, error) {
	return &MemoryUserRepository{SharedURLRows: sharedURLRows}, nil
}
