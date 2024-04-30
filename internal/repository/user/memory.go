package user

import (
	"errors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type MemoryUserRepository struct {
	Users map[int]models.User
}

func (r *MemoryUserRepository) Create() (models.User, error) {
	user := models.User{
		ID: len(r.Users) + 1,
	}
	r.Users[user.ID] = user
	return user, nil
}

func (r *MemoryUserRepository) UpdateUserToken(user models.User, token string) error {
	if _, exists := r.Users[user.ID]; !exists {
		return errors.New("user not found")
	}
	user.Token = token
	r.Users[user.ID] = user
	return nil
}

func (r *MemoryUserRepository) FindByID(userID int) (models.User, bool) {
	user, exists := r.Users[userID]
	return user, exists
}

func NewMemoryUserRepository() (*MemoryUserRepository, error) {
	return &MemoryUserRepository{Users: make(map[int]models.User)}, nil
}
