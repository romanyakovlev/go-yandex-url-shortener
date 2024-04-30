package user

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type FileUserRepository struct {
	filePath string
	Logger   *logger.Logger
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

func (r *FileUserRepository) Create() (models.User, error) {
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return models.User{}, err
	}
	defer file.Close()

	maxID := 0
	for scanner.Scan() {
		var user models.User
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &user); err != nil {
			r.Logger.Errorf("Error unmarshaling user: %v", err)
			continue
		}
		if user.ID > maxID {
			maxID = user.ID
		}
	}

	user := models.User{
		ID: maxID + 1,
	}

	data, err := json.Marshal(user)
	if err != nil {
		r.Logger.Errorf("Error marshaling user: %v", err)
		return models.User{}, err
	}

	writer, file, err := r.newWriter()
	if err != nil {
		r.Logger.Errorf("Error creating writer: %v", err)
		return models.User{}, err
	}
	defer file.Close()

	if _, err := writer.WriteString(string(data) + "\n"); err != nil {
		r.Logger.Errorf("Error writing user to file: %v", err)
		return models.User{}, err
	}
	if err := writer.Flush(); err != nil {
		r.Logger.Errorf("Error flushing writer: %v", err)
		return models.User{}, err
	}

	return user, nil
}

func (r *FileUserRepository) UpdateUserToken(user models.User, token string) error {
	scanner, file, err := r.newScanner()
	if err != nil {
		return err
	}
	defer file.Close()

	var users []models.User
	found := false

	for scanner.Scan() {
		var currentUser models.User
		if err := json.Unmarshal(scanner.Bytes(), &currentUser); err != nil {
			continue
		}
		if currentUser.ID == user.ID {
			currentUser.Token = token
			found = true
		}
		users = append(users, currentUser)
	}

	if !found {
		return errors.New("user not found")
	}
	writer, file, err := r.newWriter(true)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, u := range users {
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (r *FileUserRepository) FindByID(userID int) (models.User, bool) {
	scanner, file, err := r.newScanner()
	if err != nil {
		r.Logger.Errorf("Error creating scanner: %v", err)
		return models.User{}, false
	}
	defer file.Close()
	for scanner.Scan() {
		var user models.User
		if err := json.Unmarshal(scanner.Bytes(), &user); err != nil {
			r.Logger.Errorf("Error unmarshaling user: %v", err)
			continue
		}

		if user.ID == userID {
			return user, true
		}
	}
	if err := scanner.Err(); err != nil {
		r.Logger.Errorf("Error scanning file: %v", err)
	}

	return models.User{}, false
}

/*
func NewFileUserRepository(serverConfig config.Config, sugar *logger.Logger) (*FileUserRepository, error) {
	//fileScanner, err := NewFileScanner(serverConfig.FileStoragePath)
	fileScanner, err := NewFileScanner("./user-db.json")
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}
	//fileWriter, err := NewFileWriter(serverConfig.FileStoragePath)
	fileWriter, err := NewFileWriter("./user-db.json")
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return nil, err
	}
	return &FileUserRepository{Scanner: fileScanner, Writer: fileWriter, Logger: sugar}, nil
}

*/

func NewFileUserRepository(serverConfig config.Config, sugar *logger.Logger) (*FileUserRepository, error) {
	return &FileUserRepository{
		//filePath: serverConfig.FileStoragePath,
		filePath: "./user-db.json",
		Logger:   sugar,
	}, nil
}
