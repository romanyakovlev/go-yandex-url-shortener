package workers

import (
	"context"
	"fmt"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	shortener "github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

type DeletionRequest struct {
	User models.User
	URLs []string
}

var deletionRequestsChan = make(chan DeletionRequest, 100)

type URLDeletionWorker struct {
	shortener    *shortener.URLShortenerService
	errorChannel chan error
}

func (w *URLDeletionWorker) StartDeletionWorker(ctx context.Context) {
	for {
		select {
		case req := <-deletionRequestsChan:
			go w.processDeletionRequest(ctx, req)
		case <-ctx.Done():
			return
		}
	}
}

func (w *URLDeletionWorker) SendDeletionRequestToWorker(req DeletionRequest) error {
	select {
	case deletionRequestsChan <- req:
		return nil
	default:
		return fmt.Errorf("the deletion request queue is currently full, please try again later")
	}
}

func (w *URLDeletionWorker) processDeletionRequest(ctx context.Context, req DeletionRequest) {
	if err := w.shortener.DeleteBatchURL(req.URLs, req.User); err != nil {
		select {
		case w.errorChannel <- err:
		case <-ctx.Done():
			fmt.Println("Operation canceled, skipping error reporting.")
		}
	}
}

func (w *URLDeletionWorker) StartErrorListener(ctx context.Context) {
	for {
		select {
		case err := <-w.errorChannel:
			fmt.Printf("Error processing deletion request: %v\n", err)
		case <-ctx.Done():
			fmt.Println("Error listener shutting down due to context cancellation.")
			return
		}
	}
}

func InitURLDeletionWorker(s *shortener.URLShortenerService) *URLDeletionWorker {
	return &URLDeletionWorker{
		shortener:    s,
		errorChannel: make(chan error, 100),
	}
}
