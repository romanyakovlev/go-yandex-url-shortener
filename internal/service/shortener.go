package service

import (
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/pkg/utils"
)

type URLShortener interface {
	AddURL(urlStr string) (string, error)
	GetURL(shortURL string) (string, bool)
}

type URLShortenerService struct {
	Config config.Config
	Repo   repository.URLRepository
}

func (s URLShortenerService) AddURL(urlStr string) (string, error) {
	randomPath := utils.RandStringBytes(8)
	err := s.Repo.Save(models.URLToSave{RandomPath: randomPath, URLStr: urlStr})
	if err != nil {
		return "", err
	}
	return s.Config.BaseURL + "/" + randomPath, nil
}

func (s URLShortenerService) AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.ShortenBatchURLResponseElement, error) {
	var batchToReturn []models.ShortenBatchURLResponseElement
	var batchToSave []models.URLToSave
	for _, elem := range batchArray {
		randomPath := utils.RandStringBytes(8)
		shortURL := s.Config.BaseURL + "/" + randomPath
		batchToSave = append(batchToSave, models.URLToSave{RandomPath: randomPath, URLStr: elem.OriginalURL})
		batchToReturn = append(batchToReturn, models.ShortenBatchURLResponseElement{CorrelationID: elem.CorrelationID, ShortURL: shortURL})
	}
	err := s.Repo.BatchSave(batchToSave)
	if err != nil {
		return []models.ShortenBatchURLResponseElement{}, err
	}
	return batchToReturn, nil
}

func (s URLShortenerService) GetURL(shortURL string) (string, bool) {
	value, ok := s.Repo.Find(shortURL)
	return value, ok
}
