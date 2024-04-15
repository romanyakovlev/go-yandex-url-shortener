package service

import (
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/pkg/utils"
)

type URLRepository interface {
	Save(models.URLToSave) error
	BatchSave([]models.URLToSave) error
	Find(shortURL string) (string, bool)
	FindByOriginalURL(originalURL string) (string, bool)
}

type URLShortenerService struct {
	config config.Config
	repo   URLRepository
}

func (s URLShortenerService) AddURL(urlStr string) (string, error) {
	randomPath := utils.RandStringBytes(8)
	err := s.repo.Save(models.URLToSave{RandomPath: randomPath, URLStr: urlStr})
	if err != nil {
		return "", err
	}
	return s.config.BaseURL + "/" + randomPath, nil
}

func (s URLShortenerService) AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.ShortenBatchURLResponseElement, error) {
	var batchToReturn []models.ShortenBatchURLResponseElement
	var batchToSave []models.URLToSave
	for _, elem := range batchArray {
		randomPath := utils.RandStringBytes(8)
		shortURL := s.config.BaseURL + "/" + randomPath
		batchToSave = append(batchToSave, models.URLToSave{RandomPath: randomPath, URLStr: elem.OriginalURL})
		batchToReturn = append(batchToReturn, models.ShortenBatchURLResponseElement{CorrelationID: elem.CorrelationID, ShortURL: shortURL})
	}
	err := s.repo.BatchSave(batchToSave)
	if err != nil {
		return []models.ShortenBatchURLResponseElement{}, err
	}
	return batchToReturn, nil
}

func (s URLShortenerService) GetURL(shortURL string) (string, bool) {
	value, ok := s.repo.Find(shortURL)
	return value, ok
}

func (s URLShortenerService) GetURLByOriginalURL(originalURL string) (string, bool) {
	randomPath, ok := s.repo.FindByOriginalURL(originalURL)
	return s.config.BaseURL + "/" + randomPath, ok
}

func NewURLShortenerService(config config.Config, repo URLRepository) *URLShortenerService {
	return &URLShortenerService{
		config: config,
		repo:   repo,
	}
}
