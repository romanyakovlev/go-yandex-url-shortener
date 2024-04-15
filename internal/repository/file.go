package repository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type FileURLRepository struct {
	URLMap  map[string]string
	Scanner *bufio.Scanner
	Writer  *bufio.Writer
	Logger  *logger.Logger
}

type URLRow struct {
	UUID        uuid.UUID `json:"uuid" db:"uuid"`
	ShortURL    string    `json:"short_url" db:"short_url"`
	OriginalURL string    `json:"original_url" db:"original_url"`
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

func (r FileURLRepository) FindByOriginalURL(originalURL string) (string, bool) {
	for r.Scanner.Scan() {
		var urlRow URLRow
		line := r.Scanner.Text()
		jsonErr := json.Unmarshal([]byte(line), &urlRow)
		if jsonErr != nil {
			r.Logger.Debugf("cannot decode request JSON body: %s", jsonErr)
			return "", false
		}
		if urlRow.OriginalURL == originalURL {
			return urlRow.ShortURL, true
		}
	}
	return "", false
}

func (r FileURLRepository) Save(url models.URLToSave) error {
	URLRowObject := URLRow{UUID: uuid.New(), ShortURL: url.RandomPath, OriginalURL: url.URLStr}
	data, err := json.Marshal(URLRowObject)
	if err != nil {
		r.Logger.Debugf("Cannot encode json: %s", err)
	}
	_, err = r.Writer.WriteString(string(data) + "\n")
	if err != nil {
		r.Logger.Debugf("Cannot write data: %s", err)
		return err
	}
	r.Writer.Flush()
	return nil
}

func (r FileURLRepository) BatchSave(urls []models.URLToSave) error {
	for _, url := range urls {
		URLRowObject := URLRow{UUID: uuid.New(), ShortURL: url.RandomPath, OriginalURL: url.URLStr}
		data, err := json.Marshal(URLRowObject)
		if err != nil {
			r.Logger.Debugf("Cannot encode json: %s", err)
		}
		_, err = r.Writer.WriteString(string(data) + "\n")
		if err != nil {
			r.Logger.Debugf("Cannot write data: %s", err)
			return err
		}
		r.Writer.Flush()
	}
	return nil
}

func NewFileScanner(fileName string) (*bufio.Scanner, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return bufio.NewScanner(file), nil
}

func NewFileWriter(fileName string) (*bufio.Writer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriter(file)
	return writer, nil
}

func NewFileURLRepository(serverConfig config.Config, sugar *logger.Logger) (*FileURLRepository, error) {
	fileScanner, err := NewFileScanner(serverConfig.FileStoragePath)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}
	fileWriter, err := NewFileWriter(serverConfig.FileStoragePath)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}
	return &FileURLRepository{URLMap: make(map[string]string), Scanner: fileScanner, Writer: fileWriter, Logger: sugar}, nil
}
