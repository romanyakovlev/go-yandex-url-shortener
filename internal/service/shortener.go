package service

import (
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/pkg/utils"
)

type URLShortener interface {
	AddURL(urlStr string) string
	GetURL(shortURL string) (string, bool)
}

type URLShortenerService struct {
	Config config.Config
	Repo   repository.URLRepository
}

func (s URLShortenerService) AddURL(urlStr string) string {
	randomPath := utils.RandStringBytes(8)
	s.Repo.Save(randomPath, urlStr)
	return s.Config.BaseURL + "/" + randomPath
}

func (s URLShortenerService) GetURL(shortURL string) (string, bool) {
	value, ok := s.Repo.Find(shortURL)
	return value, ok
}
