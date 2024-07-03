package models

import (
	"sync"

	"github.com/google/uuid"
)

// ShortenURLRequest структура для запроса на сокращение URL.
type ShortenURLRequest struct {
	URL string `json:"url"` // URL для сокращения.
}

// ShortenBatchURLRequestElement элемент пакетного запроса на сокращение URL.
type ShortenBatchURLRequestElement struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для корреляции в ответе.
	OriginalURL   string `json:"original_url"`   // Исходный URL для сокращения.
}

// ShortenBatchURLResponseElement элемент пакетного ответа на сокращение URL.
type ShortenBatchURLResponseElement struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для корреляции с запросом.
	ShortURL      string `json:"short_url"`      // Сокращенный URL.
}

// ShortenURLResponse структура для ответа на сокращение URL.
type ShortenURLResponse struct {
	Result string `json:"result"` // Результат сокращения URL.
}

// URLToSave структура для сохранения URL в хранилище.
type URLToSave struct {
	RandomPath string // Случайный путь, используемый в качестве сокращенного URL.
	URLStr     string // Исходный URL.
}

// User структура пользователя.
type User struct {
	UUID  uuid.UUID // Уникальный идентификатор пользователя.
	Token string    // JWT-токен пользователя.
}

// SavedURL структура сохраненного URL.
type SavedURL struct {
	UUID     uuid.UUID // Уникальный идентификатор сохраненного URL.
	ShortURL string    // Сокращенный URL.
}

// CorrelationSavedURL структура для сохраненного URL с корреляционным идентификатором.
type CorrelationSavedURL struct {
	CorrelationID string   // Корреляционный идентификатор.
	SavedURL      SavedURL // Сохраненный URL.
}

// URLByUserResponseElement элемент ответа на запрос URL, принадлежащих пользователю.
type URLByUserResponseElement struct {
	ShortURL    string `json:"short_url"`    // Сокращенный URL.
	OriginalURL string `json:"original_url"` // Исходный URL.
}

// URLRow структура строки URL в БД
type URLRow struct {
	UUID        uuid.UUID `json:"uuid" db:"uuid"`                 // Уникальный идентификатор URL.
	ShortURL    string    `json:"short_url" db:"short_url"`       // Сокращенный URL.
	OriginalURL string    `json:"original_url" db:"original_url"` // Исходный URL.
	DeletedFlag bool      `db:"is_deleted"`                       // Флаг, указывающий на удаление URL.
	UserID      uuid.UUID `json:"user_id" db:"user_id"`           // Идентификатор пользователя, владельца URL.
}

// SharedURLRows структура для хранения и синхронизации списка URL.
type SharedURLRows struct {
	Mu      sync.Mutex // Мьютекс для синхронизации доступа к URLRows.
	URLRows []URLRow   // Список URL.
}

// NewSharedURLRows создает новый экземпляр SharedURLRows.
func NewSharedURLRows() *SharedURLRows {
	return &SharedURLRows{
		URLRows: make([]URLRow, 0),
	}
}
