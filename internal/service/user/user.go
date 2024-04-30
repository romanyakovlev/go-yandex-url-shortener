package user

import (
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/jwt"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type UserRepository interface {
	Create() (models.User, error)
	UpdateUserToken(user models.User, token string) error
	FindByID(UserID int) (models.User, bool)
}

type UserService struct {
	config config.Config
	repo   UserRepository
}

func (s UserService) CreateUserWithToken() (models.User, error) {
	user, err := s.repo.Create()
	if err != nil {
		return models.User{}, err
	}
	token, err := jwt.BuildJWTString(user.ID)
	if err != nil {
		return models.User{}, err
	}
	err = s.repo.UpdateUserToken(user, token)
	if err != nil {
		return models.User{}, err
	}
	return models.User{
		ID:    user.ID,
		Token: token,
	}, nil
}

func (s UserService) GetUserByID(ID int) (models.User, bool) {
	value, ok := s.repo.FindByID(ID)
	return value, ok
}

func NewUserService(config config.Config, repo UserRepository) *UserService {
	return &UserService{
		config: config,
		repo:   repo,
	}
}
