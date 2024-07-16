package workers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	shortener "github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

func setupURLShortenerService() (*shortener.URLShortenerService, *models.SharedURLRows) {
	sharedURLRows := models.NewSharedURLRows()
	urlRepo, _ := repository.NewMemoryURLRepository(sharedURLRows)
	userRepo, _ := repository.NewMemoryUserRepository(sharedURLRows)
	service := shortener.NewURLShortenerService(config.Config{}, urlRepo, userRepo)
	return service, sharedURLRows
}

func TestURLDeletionWorker_SendDeletionRequestToWorker(t *testing.T) {
	t.Parallel()
	service, _ := setupURLShortenerService()
	worker := InitURLDeletionWorker(service)
	req := DeletionRequest{
		User: models.User{UUID: uuid.New()},
		URLs: []string{"url1", "url2"},
	}

	err := worker.SendDeletionRequestToWorker(req)
	assert.NoError(t, err)

	for i := 0; i < cap(worker.deletionRequestsChan)-1; i++ {
		deletionErr := worker.SendDeletionRequestToWorker(req)
		assert.NoError(t, deletionErr)
	}

	err = worker.SendDeletionRequestToWorker(req)
	assert.Error(t, err, "expected an error when the deletion request queue is full")
}

func TestURLDeletionWorker_ProcessDeletionRequest(t *testing.T) {
	t.Parallel()
	service, sharedURLRows := setupURLShortenerService()
	worker := InitURLDeletionWorker(service)
	ctx := context.Background()

	userID := uuid.New()
	req := DeletionRequest{
		User: models.User{UUID: userID},
		URLs: []string{"url1", "url2"},
	}

	sharedURLRows.Mu.Lock()
	for _, url := range req.URLs {
		sharedURLRows.URLRows = append(sharedURLRows.URLRows, models.URLRow{
			UUID:        uuid.New(),
			ShortURL:    url,
			OriginalURL: "original-" + url,
			UserID:      userID,
		})
	}
	sharedURLRows.Mu.Unlock()

	go worker.processDeletionRequest(ctx, req)
	time.Sleep(100 * time.Millisecond)

	sharedURLRows.Mu.Lock()
	for _, urlRow := range sharedURLRows.URLRows {
		if urlRow.UserID == userID {
			assert.True(t, urlRow.DeletedFlag, "expected url to be marked as deleted")
		}
	}
	sharedURLRows.Mu.Unlock()
}

func TestURLDeletionWorker_StartDeletionWorker(t *testing.T) {
	t.Parallel()
	service, sharedURLRows := setupURLShortenerService()
	worker := InitURLDeletionWorker(service)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userID := uuid.New()
	req := DeletionRequest{
		User: models.User{UUID: userID},
		URLs: []string{"url1", "url2"},
	}

	sharedURLRows.Mu.Lock()
	for _, url := range req.URLs {
		sharedURLRows.URLRows = append(sharedURLRows.URLRows, models.URLRow{
			UUID:        uuid.New(),
			ShortURL:    url,
			OriginalURL: "original-" + url,
			UserID:      userID,
		})
	}
	sharedURLRows.Mu.Unlock()

	go worker.StartDeletionWorker(ctx)

	err := worker.SendDeletionRequestToWorker(req)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	sharedURLRows.Mu.Lock()
	for _, urlRow := range sharedURLRows.URLRows {
		if urlRow.UserID == userID {
			assert.True(t, urlRow.DeletedFlag, "expected url to be marked as deleted")
		}
	}
	sharedURLRows.Mu.Unlock()
}
