package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

// FileURLRepository представляет репозиторий URL, хранящийся в файле.
type FileURLRepository struct {
	filePath string         // Путь к файлу для хранения данных.
	Logger   *logger.Logger // Логгер для регистрации событий.
}

// FileUserRepository представляет репозиторий пользователей, хранящийся в файле.
type FileUserRepository struct {
	filePath string         // Путь к файлу для хранения данных.
	Logger   *logger.Logger // Логгер для регистрации событий.
}

// newScanner создает новый сканер для чтения данных из файла.
func (r *FileURLRepository) newScanner() (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewScanner(file), file, nil
}

// newWriter создает новый writer для записи данных в файл.
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

// Find ищет URL по сокращенному адресу в файле.
func (r FileURLRepository) Find(shortURL string) (models.URLRow, bool) {
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return models.URLRow{}, false
	}
	defer file.Close()
	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
		err := json.Unmarshal([]byte(line), &urlRow)
		if err != nil {
			r.Logger.Debugf("cannot decode request JSON body: %s", err)
			return models.URLRow{}, false
		}
		if urlRow.ShortURL == shortURL {
			return urlRow, true
		}
	}
	return models.URLRow{}, false
}

// FindByOriginalURL ищет сокращенный URL по оригинальному адресу в файле.
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
		err := json.Unmarshal([]byte(line), &urlRow)
		if err != nil {
			r.Logger.Debugf("cannot decode request JSON body: %s", err)
			return "", false
		}
		if urlRow.OriginalURL == originalURL {
			return urlRow.ShortURL, true
		}
	}
	return "", false
}

// FindByUserID ищет все URL, принадлежащие пользователю, в файле.
func (r *FileURLRepository) FindByUserID(userID uuid.UUID) ([]models.URLRow, bool) {
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
		err := json.Unmarshal([]byte(line), &urlRow)
		if err != nil {
			r.Logger.Debugf("cannot decode line JSON: %s", err)
			continue
		}
		if urlRow.UserID == userID {
			urlRows = append(urlRows, urlRow)
		}
	}

	return urlRows, true
}

// Save сохраняет новый URL в файл.
func (r FileURLRepository) Save(url models.URLToSave) (uuid.UUID, error) {
	writer, file, err := r.newWriter()
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return uuid.UUID{}, err
	}
	defer file.Close()

	UUID := uuid.New()
	URLRowObject := models.URLRow{UUID: UUID, ShortURL: url.RandomPath, OriginalURL: url.URLStr, DeletedFlag: false}
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

// BatchSave сохраняет несколько URL в файл одной транзакцией.
func (r FileURLRepository) BatchSave(urls []models.URLToSave) ([]uuid.UUID, error) {
	writer, file, err := r.newWriter()
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return []uuid.UUID{}, err
	}
	defer file.Close()

	var UUIDs []uuid.UUID
	var errs []error

	for _, url := range urls {
		UUID := uuid.New()
		URLRowObject := models.URLRow{UUID: UUID, ShortURL: url.RandomPath, OriginalURL: url.URLStr}
		UUIDs = append(UUIDs, URLRowObject.UUID)
		data, err := json.Marshal(URLRowObject)
		if err != nil {
			r.Logger.Debugf("Cannot encode json: %s", err)
			errs = append(errs, err)
			continue
		}
		_, err = writer.WriteString(string(data) + "\n")
		if err != nil {
			r.Logger.Debugf("Cannot write data: %s", err)
			errs = append(errs, err)
			continue
		}
		if err := writer.Flush(); err != nil {
			r.Logger.Errorf("Error flushing writer: %v", err)
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return UUIDs, errors.Join(errs...)
	}

	return UUIDs, nil
}

// BatchDelete помечает URL как удаленные для указанного пользователя в файле.
func (r *FileURLRepository) BatchDelete(urls []string, userID uuid.UUID) error {
	uuidMap := make(map[string]bool)
	for _, shortURL := range urls {
		uuidMap[shortURL] = true
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

		if _, exists := uuidMap[urlRow.ShortURL]; exists {
			urlRow.DeletedFlag = true
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

	var errs []error

	for _, urlRow := range urlRows {
		data, err := json.Marshal(urlRow)
		if err != nil {
			r.Logger.Debugf("Cannot encode URLRow to JSON: %s", err)
			errs = append(errs, err)
			continue
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			r.Logger.Debugf("Cannot write URLRow to file: %s", err)
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if err := writer.Flush(); err != nil {
		r.Logger.Debugf("Error flushing writer: %s", err)
		return err
	}

	return nil
}

func (r *FileUserRepository) newScanner() (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewScanner(file), file, nil
}

func (r *FileUserRepository) newWriter(truncateModes ...bool) (*bufio.Writer, *os.File, error) {
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

// UpdateUser обновляет пользователя для указанного URL.
func (r *FileUserRepository) UpdateUser(savedURLUUID uuid.UUID, userID uuid.UUID) error {
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
			urlRow.UUID = userID
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

	var errs []error

	for _, urlRow := range urlRows {
		data, err := json.Marshal(urlRow)
		if err != nil {
			r.Logger.Debugf("Cannot encode URLRow to JSON: %s", err)
			errs = append(errs, err)
			continue
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			r.Logger.Debugf("Cannot write URLRow to file: %s", err)
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if err := writer.Flush(); err != nil {
		r.Logger.Debugf("Error flushing writer: %s", err)
		return err
	}

	return nil
}

// UpdateBatchUser обновляет пользователя для нескольких URL.
func (r *FileUserRepository) UpdateBatchUser(savedURLUUIDs []uuid.UUID, userID uuid.UUID) error {
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

	var errs []error

	for _, urlRow := range urlRows {
		data, err := json.Marshal(urlRow)
		if err != nil {
			r.Logger.Debugf("Cannot encode URLRow to JSON: %s", err)
			errs = append(errs, err)
			continue
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			r.Logger.Debugf("Cannot write URLRow to file: %s", err)
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if err := writer.Flush(); err != nil {
		r.Logger.Debugf("Error flushing writer: %s", err)
		return err
	}

	return nil
}

// GetStats возвращает статистику
func (r *FileURLRepository) GetStats() (models.URLStats, bool) {
	var stats models.URLStats
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return models.URLStats{}, false
	}
	defer file.Close()

	uniqueUsers := make(map[uuid.UUID]struct{})
	uniqueURLs := make(map[string]struct{})

	for scanner.Scan() {
		var urlRow models.URLRow
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &urlRow); err != nil {
			r.Logger.Debugf("Cannot decode line JSON: %s", err)
			continue
		}

		if _, exists := uniqueURLs[urlRow.ShortURL]; !exists {
			uniqueURLs[urlRow.ShortURL] = struct{}{}
			stats.URLs++
		}

		if _, exists := uniqueUsers[urlRow.UserID]; !exists {
			uniqueUsers[urlRow.UserID] = struct{}{}
			stats.Users++
		}
	}

	return stats, true
}

// NewFileURLRepository создает новый экземпляр репозитория URL, хранящегося в файле.
func NewFileURLRepository(serverConfig config.Config, sugar *logger.Logger) (*FileURLRepository, error) {
	return &FileURLRepository{filePath: serverConfig.FileStoragePath, Logger: sugar}, nil
}

// NewFileUserRepository создает новый экземпляр репозитория пользователей, хранящегося в файле.
func NewFileUserRepository(serverConfig config.Config, sugar *logger.Logger) (*FileUserRepository, error) {
	return &FileUserRepository{filePath: serverConfig.FileStoragePath, Logger: sugar}, nil
}
