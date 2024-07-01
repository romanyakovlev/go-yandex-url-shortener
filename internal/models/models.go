package models

import (
	"sync"

	"github.com/google/uuid"
)

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenBatchURLRequestElement struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchURLResponseElement struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ShortenURLResponse struct {
	Result string `json:"result"`
}

type URLToSave struct {
	RandomPath string
	URLStr     string
}

type User struct {
	UUID  uuid.UUID
	Token string
}

type SavedURL struct {
	UUID     uuid.UUID
	ShortURL string
}

type CorrelationSavedURL struct {
	CorrelationID string
	SavedURL      SavedURL
}

type URLByUserResponseElement struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLRow struct {
	UUID        uuid.UUID `json:"uuid" db:"uuid"`
	ShortURL    string    `json:"short_url" db:"short_url"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	DeletedFlag bool      `db:"is_deleted"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
}

type SharedURLRows struct {
	Mu      sync.Mutex // Embed a mutex directly into the struct
	URLRows []URLRow
}

func NewSharedURLRows() *SharedURLRows {
	return &SharedURLRows{
		URLRows: make([]URLRow, 0),
	}
}
