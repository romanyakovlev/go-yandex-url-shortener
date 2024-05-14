package service

import (
	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/pkg/utils"
)

type URLRepository interface {
	Save(models.URLToSave) (uuid.UUID, error)
	BatchSave([]models.URLToSave) ([]uuid.UUID, error)
	BatchDelete(urls []string, userID uuid.UUID) error
	Find(shortURL string) (models.URLRow, bool)
	FindByUserID(userID uuid.UUID) ([]models.URLRow, bool)
	FindByOriginalURL(originalURL string) (string, bool)
}

type UserRepository interface {
	UpdateUser(SavedURLUUID uuid.UUID, userID uuid.UUID) error
	UpdateBatchUser(SavedURLUUIDs []uuid.UUID, userID uuid.UUID) error
}

type URLShortenerService struct {
	config   config.Config
	urlRepo  URLRepository
	userRepo UserRepository
}

func (s URLShortenerService) AddURL(urlStr string) (models.SavedURL, error) {
	randomPath := utils.RandStringBytes(8)
	UUID, err := s.urlRepo.Save(models.URLToSave{RandomPath: randomPath, URLStr: urlStr})
	if err != nil {
		return models.SavedURL{}, err
	}
	return models.SavedURL{UUID: UUID, ShortURL: s.config.BaseURL + "/" + randomPath}, nil
}

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

func (s URLShortenerService) AddUserToURL(SavedURL models.SavedURL, user models.User) error {
	err := s.userRepo.UpdateUser(SavedURL.UUID, user.UUID)
	if err != nil {
		return err
	}
	return nil
}

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

func (s URLShortenerService) GetURL(shortURL string) (models.URLRow, bool) {
	row, ok := s.urlRepo.Find(shortURL)
	return row, ok
}

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

func (s URLShortenerService) GetURLByOriginalURL(originalURL string) (string, bool) {
	randomPath, ok := s.urlRepo.FindByOriginalURL(originalURL)
	return s.config.BaseURL + "/" + randomPath, ok
}

func (s URLShortenerService) DeleteBatchURL(urls []string, user models.User) error {
	err := s.urlRepo.BatchDelete(urls, user.UUID)
	return err
}

func (s URLShortenerService) ConvertCorrelationSavedURLToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement {
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

func (s URLShortenerService) ConvertCorrelationSavedURLToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL {
	var elements []models.SavedURL

	for _, item := range correlationSavedURLs {

		elements = append(elements, item.SavedURL)
	}

	return elements
}

func NewURLShortenerService(config config.Config, urlRepo URLRepository, userRepo UserRepository) *URLShortenerService {
	return &URLShortenerService{
		config:   config,
		urlRepo:  urlRepo,
		userRepo: userRepo,
	}
}
