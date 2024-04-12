package repository

import "github.com/romanyakovlev/go-yandex-url-shortener/internal/models"

type MemoryURLRepository struct {
	URLMap map[string]string
}

func (r MemoryURLRepository) Save(url models.URLToSave) error {
	r.URLMap[url.RandomPath] = url.URLStr
	return nil
}

func (r MemoryURLRepository) BatchSave(urls []models.URLToSave) error {
	for _, url := range urls {
		r.URLMap[url.RandomPath] = url.URLStr
	}
	return nil
}

func (r MemoryURLRepository) Find(shortURL string) (string, bool) {
	value, ok := r.URLMap[shortURL]
	return value, ok
}
