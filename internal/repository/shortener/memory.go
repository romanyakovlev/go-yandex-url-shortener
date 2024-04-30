package shortener

import (
	"errors"
	"github.com/google/uuid"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type MemoryURLRepository struct {
	URLRows []models.URLRow
}

func (r *MemoryURLRepository) Save(url models.URLToSave) (uuid.UUID, error) {
	UUID := uuid.New()
	newURLRow := models.URLRow{
		UUID:        UUID,
		ShortURL:    url.RandomPath,
		OriginalURL: url.URLStr,
	}
	r.URLRows = append(r.URLRows, newURLRow)
	return UUID, nil
}

func (r *MemoryURLRepository) BatchSave(urls []models.URLToSave) ([]uuid.UUID, error) {
	var UUIDs []uuid.UUID
	for _, url := range urls {
		UUID := uuid.New()
		newURLRow := models.URLRow{
			UUID:        UUID,
			ShortURL:    url.RandomPath,
			OriginalURL: url.URLStr,
		}
		UUIDs = append(UUIDs, UUID)
		r.URLRows = append(r.URLRows, newURLRow)
	}
	return UUIDs, nil
}

func (r *MemoryURLRepository) Find(shortURL string) (string, bool) {
	for _, urlRow := range r.URLRows {
		if urlRow.ShortURL == shortURL {
			return urlRow.OriginalURL, true
		}
	}
	return "", false
}

func (r *MemoryURLRepository) FindByOriginalURL(originalURL string) (string, bool) {
	for _, urlRow := range r.URLRows {
		if urlRow.OriginalURL == originalURL {
			return urlRow.ShortURL, true
		}
	}
	return "", false
}

func (r *MemoryURLRepository) UpdateUser(SavedURLUUID uuid.UUID, userID int) error {
	found := false
	for i, urlRow := range r.URLRows {
		if urlRow.UUID == SavedURLUUID {
			r.URLRows[i].UserID = userID
			found = true
			break
		}
	}
	if !found {
		return errors.New("URL not found")
	}
	return nil
}

func (r *MemoryURLRepository) FindByUserID(userID int) ([]models.URLRow, bool) {
	var matchedURLs []models.URLRow
	found := false
	for _, urlRow := range r.URLRows {
		if urlRow.UserID == userID {
			matchedURLs = append(matchedURLs, urlRow)
			found = true
		}
	}
	return matchedURLs, found
}

func NewMemoryURLRepository() (*MemoryURLRepository, error) {
	return &MemoryURLRepository{URLRows: make([]models.URLRow, 0)}, nil
}

func (r *MemoryURLRepository) UpdateBatchUser(SavedURLUUIDs []uuid.UUID, userID int) error {
	uuidMap := make(map[uuid.UUID]bool)
	for _, id := range SavedURLUUIDs {
		uuidMap[id] = true
	}

	updated := false
	for i, urlRow := range r.URLRows {
		if _, exists := uuidMap[urlRow.UUID]; exists {
			r.URLRows[i].UserID = userID
			updated = true
		}
	}

	if !updated {
		return errors.New("no URLs updated")
	}
	return nil
}
