// Package workers выполняет роль фонового процесса, который
// выполняет отложенное удаление URL отдельно от хендлера.
package workers

import (
	"context"
	"fmt"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	shortener "github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

// DeletionRequest структура запроса на удаление URL.
type DeletionRequest struct {
	User models.User // Пользователь, от имени которого производится удаление.
	URLs []string    // Список URL для удаления.
}

// URLDeletionWorker структура фонового процесса для удаления URL.
type URLDeletionWorker struct {
	shortener            *shortener.URLShortenerService // Сервис сокращения URL.
	errorChannel         chan error                     // Канал для передачи ошибок.
	deletionRequestsChan chan DeletionRequest           // Канал для запросов на удаление.
}

// StartDeletionWorker запускает фоновый процесс для обработки запросов на удаление.
func (w *URLDeletionWorker) StartDeletionWorker(ctx context.Context) {
	for {
		select {
		case req := <-w.deletionRequestsChan: // Чтение запроса на удаление из канала.
			go w.processDeletionRequest(ctx, req) // Асинхронная обработка запроса.
		case <-ctx.Done(): // Завершение работы при отмене контекста.
			return
		}
	}
}

// SendDeletionRequestToWorker отправляет запрос на удаление в фоновый процесс.
func (w *URLDeletionWorker) SendDeletionRequestToWorker(req DeletionRequest) error {
	select {
	case w.deletionRequestsChan <- req: // Попытка отправить запрос в канал.
		return nil
	default:
		return fmt.Errorf("the deletion request queue is currently full, please try again later")
	}
}

// processDeletionRequest обрабатывает запрос на удаление.
func (w *URLDeletionWorker) processDeletionRequest(ctx context.Context, req DeletionRequest) {
	if err := w.shortener.DeleteBatchURL(req.URLs, req.User); err != nil {
		select {
		case w.errorChannel <- err:
		case <-ctx.Done():
			fmt.Println("Operation canceled, skipping error reporting.")
		}
	}
}

// StartErrorListener запускает прослушивание канала ошибок.
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

// InitURLDeletionWorker инициализирует и возвращает новый экземпляр фонового процесса для удаления URL.
func InitURLDeletionWorker(s *shortener.URLShortenerService) *URLDeletionWorker {
	return &URLDeletionWorker{
		shortener:            s,
		errorChannel:         make(chan error, 100),
		deletionRequestsChan: make(chan DeletionRequest, 100),
	}
}
