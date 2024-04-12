package repository

import "github.com/romanyakovlev/go-yandex-url-shortener/internal/models"

type URLRepository interface {
	Save(models.URLToSave) error
	BatchSave([]models.URLToSave) error
	Find(shortURL string) (string, bool)
}
