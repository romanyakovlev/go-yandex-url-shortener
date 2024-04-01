package repository

import (
	"bufio"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

type FileURLRepository struct {
	URLMap  map[string]string
	Scanner *bufio.Scanner
	Writer  *bufio.Writer
	Logger  *logger.Logger
}

type URLRow struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

func (r FileURLRepository) Find(shortURL string) (string, bool) {
	for r.Scanner.Scan() {
		var urlRow URLRow
		line := r.Scanner.Text()
		jsonErr := json.Unmarshal([]byte(line), &urlRow)
		if jsonErr != nil {
			r.Logger.Debugf("cannot decode request JSON body: %s", jsonErr)
			return "", false
		}
		if urlRow.ShortURL == shortURL {
			return urlRow.OriginalURL, true
		}
	}
	return "", false
}

func (r FileURLRepository) Save(randomPath string, urlStr string) {
	URLRowObject := URLRow{UUID: uuid.New(), ShortURL: randomPath, OriginalURL: urlStr}
	data, err := json.Marshal(URLRowObject)
	if err != nil {
		r.Logger.Debugf("Cannot encode json: %s", err)
	}
	_, err = r.Writer.WriteString(string(data) + "\n")
	if err != nil {
		r.Logger.Debugf("Cannot write data: %s", err)
	}
	r.Writer.Flush()
}
