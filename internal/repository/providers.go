package repository

import (
	"bufio"
	"database/sql"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"os"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

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

func NewURLRepository(serverConfig config.Config, DB *sql.DB, sugar *logger.Logger) (URLRepository, error) {
	if serverConfig.DatabaseDSN != "" {
		return DBURLRepository{DB: DB}, nil
	} else if serverConfig.FileStoragePath != "" {
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
		return FileURLRepository{URLMap: make(map[string]string), Scanner: fileScanner, Writer: fileWriter, Logger: sugar}, nil
	} else {
		return MemoryURLRepository{URLMap: make(map[string]string)}, nil
	}
}
