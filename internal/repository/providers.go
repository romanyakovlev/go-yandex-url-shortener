package repository

import (
	"bufio"
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

func NewURLRepository(fileName string, sugar *logger.Logger) (URLRepository, error) {
	if fileName != "" {
		fileScanner, err := NewFileScanner(fileName)
		if err != nil {
			sugar.Errorf("Server error: %v", err)
			return nil, err
		}
		fileWriter, err := NewFileWriter(fileName)
		if err != nil {
			sugar.Errorf("Server error: %v", err)
			return nil, err
		}
		return FileURLRepository{URLMap: make(map[string]string), Scanner: fileScanner, Writer: fileWriter, Logger: sugar}, nil
	} else {
		return MemoryURLRepository{URLMap: make(map[string]string)}, nil
	}
}
