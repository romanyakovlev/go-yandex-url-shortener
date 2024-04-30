package shortener

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
	filePath string
	Logger   *logger.Logger
}

func (r *FileURLRepository) newScanner() (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewScanner(file), file, nil
}

func (r *FileURLRepository) newWriter(truncateModes ...bool) (*bufio.Writer, *os.File, error) {
	truncateMode := false
	if len(truncateModes) > 0 {
		truncateMode = truncateModes[0]
	}

	flags := os.O_WRONLY | os.O_CREATE
	if truncateMode {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(r.filePath, flags, 0666)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewWriter(file), file, nil
}

func (r FileURLRepository) Find(shortURL string) (string, bool) {
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return "", false
	}
	defer file.Close()
	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
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
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return "", false
	}
	defer file.Close()
	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
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

func (r *FileURLRepository) FindByUserID(userID int) ([]models.URLRow, bool) {
	var urlRows []models.URLRow
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return []models.URLRow{}, false
	}
	defer file.Close()

	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
		jsonErr := json.Unmarshal([]byte(line), &urlRow)
		if jsonErr != nil {
			r.Logger.Debugf("cannot decode line JSON: %s", jsonErr)
			continue
		}
		if urlRow.UserID == userID {
			urlRows = append(urlRows, urlRow)
		}
	}

	return urlRows, true
}

func (r FileURLRepository) Save(url models.URLToSave) (uuid.UUID, error) {
	writer, file, err := r.newWriter()
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return uuid.UUID{}, err
	}
	defer file.Close()

	UUID := uuid.New()
	URLRowObject := models.URLRow{UUID: UUID, ShortURL: url.RandomPath, OriginalURL: url.URLStr}
	data, err := json.Marshal(URLRowObject)
	if err != nil {
		r.Logger.Debugf("Cannot encode json: %s", err)
	}
	_, err = writer.WriteString(string(data) + "\n")
	if err != nil {
		r.Logger.Debugf("Cannot write data: %s", err)
		return UUID, err
	}
	if err := writer.Flush(); err != nil {
		r.Logger.Errorf("Error flushing writer: %v", err)
		return uuid.UUID{}, err
	}
	return UUID, nil
}

func (r FileURLRepository) BatchSave(urls []models.URLToSave) ([]uuid.UUID, error) {
	writer, file, err := r.newWriter()
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return []uuid.UUID{}, err
	}
	defer file.Close()

	var UUIDs []uuid.UUID
	for _, url := range urls {
		UUID := uuid.New()
		URLRowObject := models.URLRow{UUID: UUID, ShortURL: url.RandomPath, OriginalURL: url.URLStr}
		UUIDs = append(UUIDs, URLRowObject.UUID)
		data, err := json.Marshal(URLRowObject)
		if err != nil {
			r.Logger.Debugf("Cannot encode json: %s", err)
		}
		_, err = writer.WriteString(string(data) + "\n")
		if err != nil {
			r.Logger.Debugf("Cannot write data: %s", err)
			return UUIDs, err
		}
		if err := writer.Flush(); err != nil {
			r.Logger.Errorf("Error flushing writer: %v", err)
			return []uuid.UUID{}, err
		}
	}
	return UUIDs, nil
}

func (r *FileURLRepository) UpdateUser(savedURLUUID uuid.UUID, userID int) error {
	var urlRows []models.URLRow
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return err
	}
	defer file.Close()
	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &urlRow); err != nil {
			r.Logger.Debugf("Cannot decode line JSON: %s", err)
			continue
		}

		if urlRow.UUID == savedURLUUID {
			urlRow.UserID = userID
		}

		urlRows = append(urlRows, urlRow)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	writer, file, err := r.newWriter(true)
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return err
	}
	defer file.Close()
	for _, urlRow := range urlRows {
		data, err := json.Marshal(urlRow)
		if err != nil {
			r.Logger.Debugf("Cannot encode URLRow to JSON: %s", err)
			return err
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			r.Logger.Debugf("Cannot write URLRow to file: %s", err)
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		r.Logger.Debugf("Error flushing writer: %s", err)
		return err
	}

	return nil
}

func (r *FileURLRepository) UpdateBatchUser(savedURLUUIDs []uuid.UUID, userID int) error {
	uuidMap := make(map[uuid.UUID]bool)
	for _, id := range savedURLUUIDs {
		uuidMap[id] = true
	}

	var urlRows []models.URLRow
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return err
	}
	defer file.Close()
	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &urlRow); err != nil {
			r.Logger.Debugf("Cannot decode line JSON: %s", err)
			continue
		}

		if _, exists := uuidMap[urlRow.UUID]; exists {
			urlRow.UserID = userID
		}

		urlRows = append(urlRows, urlRow)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	writer, file, err := r.newWriter(true)
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return err
	}
	defer file.Close()
	for _, urlRow := range urlRows {
		data, err := json.Marshal(urlRow)
		if err != nil {
			r.Logger.Debugf("Cannot encode URLRow to JSON: %s", err)
			return err
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			r.Logger.Debugf("Cannot write URLRow to file: %s", err)
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		r.Logger.Debugf("Error flushing writer: %s", err)
		return err
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
	/*
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

	*/
	return &FileURLRepository{filePath: serverConfig.FileStoragePath, Logger: sugar}, nil
}
