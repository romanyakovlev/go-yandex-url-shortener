package models

import "github.com/google/uuid"

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
	ID    int
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
	UserID      int       `json:"user_id" db:"user_id"`
}
