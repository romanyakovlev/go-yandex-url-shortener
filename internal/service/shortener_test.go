package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
)

func setupURLShortenerService() (*URLShortenerService, *models.SharedURLRows) {
	sharedURLRows := models.NewSharedURLRows() // Assumes NewSharedURLRows initializes a mutex.
	urlRepo, _ := repository.NewMemoryURLRepository(sharedURLRows)
	userRepo, _ := repository.NewMemoryUserRepository(sharedURLRows)
	service := NewURLShortenerService(config.Config{BaseURL: "http://localhost:8000"}, urlRepo, userRepo)
	return service, sharedURLRows
}

func TestAddURL(t *testing.T) {
	service, _ := setupURLShortenerService()

	originalURL := "http://practicum.yandex.ru/example"
	savedURL, err := service.AddURL(originalURL)

	assert.NoError(t, err)
	assert.Contains(t, savedURL.ShortURL, service.config.BaseURL)
}

func TestGetURL(t *testing.T) {
	service, _ := setupURLShortenerService()

	originalURL := "http://practicum.yandex.ru/example"
	savedURL, err := service.AddURL(originalURL)
	assert.NoError(t, err)

	foundURL, found := service.GetURL(savedURL.ShortURL[len(savedURL.ShortURL)-8:])
	assert.True(t, found)
	assert.Equal(t, originalURL, foundURL.OriginalURL)
}

func TestAddBatchURL(t *testing.T) {
	service, _ := setupURLShortenerService()

	batchArray := []models.ShortenBatchURLRequestElement{
		{CorrelationID: "1", OriginalURL: "http://practicum.yandex.ru/example1"},
		{CorrelationID: "2", OriginalURL: "http://practicum.yandex.ru/example2"},
	}

	batchToReturn, err := service.AddBatchURL(batchArray)
	assert.NoError(t, err)
	assert.Equal(t, len(batchArray), len(batchToReturn))

	for i, elem := range batchArray {
		assert.Equal(t, elem.CorrelationID, batchToReturn[i].CorrelationID)
		assert.Contains(t, batchToReturn[i].SavedURL.ShortURL, service.config.BaseURL)
	}
}

func TestAddUserToURL(t *testing.T) {
	service, _ := setupURLShortenerService()

	originalURL := "http://practicum.yandex.ru/example"
	savedURL, err := service.AddURL(originalURL)
	assert.NoError(t, err)

	user := models.User{UUID: uuid.New()}
	err = service.AddUserToURL(savedURL, user)
	assert.NoError(t, err)

	urlRow, found := service.urlRepo.Find(savedURL.ShortURL[len(savedURL.ShortURL)-8:])
	assert.True(t, found)
	assert.Equal(t, user.UUID, urlRow.UserID)
}

func TestDeleteBatchURL(t *testing.T) {
	service, _ := setupURLShortenerService()

	batchArray := []models.ShortenBatchURLRequestElement{
		{CorrelationID: "1", OriginalURL: "http://practicum.yandex.ru/example1"},
		{CorrelationID: "2", OriginalURL: "http://practicum.yandex.ru/example2"},
	}

	batchToReturn, err := service.AddBatchURL(batchArray)
	assert.NoError(t, err)

	var shortURLs []string
	for _, savedURL := range batchToReturn {
		shortURLs = append(shortURLs, savedURL.SavedURL.ShortURL[len(savedURL.SavedURL.ShortURL)-8:])
	}

	user := models.User{UUID: uuid.New()}
	err = service.DeleteBatchURL(shortURLs, user)
	assert.NoError(t, err)
	/*
		for _, shortURL := range shortURLs {
			urlRow, found := service.urlRepo.Find(shortURL)
			assert.True(t, found)
			assert.True(t, urlRow.DeletedFlag)
		}

	*/
}
