// Package service содержит бизнес-логику по управлению сервисом сокращения ссылок.
package service

import (
	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/pkg/utils"
)

// URLRepository определяет интерфейс для работы с хранилищем URL.
type URLRepository interface {
	Save(models.URLToSave) (uuid.UUID, error)              // Save сохраняет URL.
	BatchSave([]models.URLToSave) ([]uuid.UUID, error)     // BatchSave сохраняет список URL.
	BatchDelete(urls []string, userID uuid.UUID) error     // BatchDelete удаляет список URL.
	Find(shortURL string) (models.URLRow, bool)            // Find выполняет поиск URL по короткому адресу.
	FindByUserID(userID uuid.UUID) ([]models.URLRow, bool) // FindByUserID ищет все URL, принадлежащие пользователю.
	FindByOriginalURL(originalURL string) (string, bool)   // FindByOriginalURL ищет URL по оригинальному адресу.
	GetStats() (models.URLStats, bool)                     // GetStats возвращает статистику
}

// UserRepository определяет интерфейс для работы с хранилищем пользователей.
type UserRepository interface {
	UpdateUser(SavedURLUUID uuid.UUID, userID uuid.UUID) error         // UpdateUser привязывает URL к пользователю.
	UpdateBatchUser(SavedURLUUIDs []uuid.UUID, userID uuid.UUID) error // UpdateBatchUser привязывает список URL к пользователю.
}

// URLShortenerService предоставляет методы для работы с сокращением URL.
type URLShortenerService struct {
	config   config.Config  // Конфигурация сервиса.
	urlRepo  URLRepository  // Репозиторий для работы с URL.
	userRepo UserRepository // Репозиторий для работы с пользователями.
}

// AddURL сокращает одиночный URL.
func (s URLShortenerService) AddURL(urlStr string) (models.SavedURL, error) {
	randomPath := utils.RandStringBytes(8)
	UUID, err := s.urlRepo.Save(models.URLToSave{RandomPath: randomPath, URLStr: urlStr})
	if err != nil {
		return models.SavedURL{}, err
	}
	return models.SavedURL{UUID: UUID, ShortURL: s.config.BaseURL + "/" + randomPath}, nil
}

// AddBatchURL сокращает список URL.
func (s URLShortenerService) AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.CorrelationSavedURL, error) {
	var batchToSave []models.URLToSave
	for _, elem := range batchArray {
		randomPath := utils.RandStringBytes(8)
		batchToSave = append(batchToSave, models.URLToSave{RandomPath: randomPath, URLStr: elem.OriginalURL})
	}

	UUIDs, err := s.urlRepo.BatchSave(batchToSave)
	if err != nil {
		return nil, err
	}

	var batchToReturn []models.CorrelationSavedURL
	for i, UUID := range UUIDs {
		shortURL := s.config.BaseURL + "/" + batchToSave[i].RandomPath
		batchToReturn = append(batchToReturn, models.CorrelationSavedURL{
			CorrelationID: batchArray[i].CorrelationID,
			SavedURL: models.SavedURL{
				UUID:     UUID,
				ShortURL: shortURL,
			},
		})
	}

	return batchToReturn, nil
}

// AddUserToURL привязывает URL к пользователю.
func (s URLShortenerService) AddUserToURL(SavedURL models.SavedURL, user models.User) error {
	err := s.userRepo.UpdateUser(SavedURL.UUID, user.UUID)
	if err != nil {
		return err
	}
	return nil
}

// AddBatchUserToURL привязывает список URL к пользователю.
func (s URLShortenerService) AddBatchUserToURL(SavedURLs []models.SavedURL, user models.User) error {
	var UUIDs []uuid.UUID
	for _, savedURL := range SavedURLs {
		UUIDs = append(UUIDs, savedURL.UUID)
	}

	err := s.userRepo.UpdateBatchUser(UUIDs, user.UUID)
	if err != nil {
		return err
	}
	return nil
}

// GetURL возвращает оригинальный URL по сокращенному адресу.
func (s URLShortenerService) GetURL(shortURL string) (models.URLRow, bool) {
	row, ok := s.urlRepo.Find(shortURL)
	return row, ok
}

// GetURLByUser возвращает список URL, принадлежащих пользователю.
func (s URLShortenerService) GetURLByUser(user models.User) ([]models.URLByUserResponseElement, bool) {
	respElements := []models.URLByUserResponseElement{}
	URLRows, ok := s.urlRepo.FindByUserID(user.UUID)
	for _, URLRow := range URLRows {
		respElements = append(respElements, models.URLByUserResponseElement{
			ShortURL:    s.config.BaseURL + "/" + URLRow.ShortURL,
			OriginalURL: URLRow.OriginalURL,
		})
	}
	return respElements, ok
}

// GetURLByOriginalURL возвращает сокращенный URL по оригинальному адресу.
func (s URLShortenerService) GetURLByOriginalURL(originalURL string) (string, bool) {
	randomPath, ok := s.urlRepo.FindByOriginalURL(originalURL)
	return s.config.BaseURL + "/" + randomPath, ok
}

// DeleteBatchURL удаляет список URL, принадлежащих пользователю.
func (s URLShortenerService) DeleteBatchURL(urls []string, user models.User) error {
	err := s.urlRepo.BatchDelete(urls, user.UUID)
	return err
}

// GetStats возвращает статистику
func (s URLShortenerService) GetStats() (models.URLStats, bool) {
	row, ok := s.urlRepo.GetStats()
	return row, ok
}

// ConvertCorrelationSavedURLsToResponse конвертирует сохраненные URL с корреляционными идентификаторами в формат ответа.
func (s URLShortenerService) ConvertCorrelationSavedURLsToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement {
	var responseElements []models.ShortenBatchURLResponseElement

	for _, item := range correlationSavedURLs {
		responseElement := models.ShortenBatchURLResponseElement{
			CorrelationID: item.CorrelationID,
			ShortURL:      item.SavedURL.ShortURL,
		}
		responseElements = append(responseElements, responseElement)
	}

	return responseElements
}

// ConvertCorrelationSavedURLsToSavedURL конвертирует сохраненные URL с корреляционными идентификаторами в сохраненные URL.
func (s URLShortenerService) ConvertCorrelationSavedURLsToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL {
	var elements []models.SavedURL

	for _, item := range correlationSavedURLs {
		elements = append(elements, item.SavedURL)
	}

	return elements
}

// NewURLShortenerService создает новый экземпляр сервиса сокращения URL.
func NewURLShortenerService(config config.Config, urlRepo URLRepository, userRepo UserRepository) *URLShortenerService {
	return &URLShortenerService{
		config:   config,
		urlRepo:  urlRepo,
		userRepo: userRepo,
	}
}
