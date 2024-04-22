package models

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
