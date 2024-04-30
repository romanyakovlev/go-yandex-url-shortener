package user

import (
	"database/sql"
	"fmt"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type DBUserRepository struct {
	db *sql.DB
}

func (r DBUserRepository) Create() (models.User, error) {
	var user models.User
	query := "INSERT INTO users DEFAULT VALUES RETURNING id"
	err := r.db.QueryRow(query).Scan(&user.ID)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r DBUserRepository) UpdateUserToken(user models.User, token string) error {
	query := "UPDATE users SET token = $1 WHERE id = $2"
	_, err := r.db.Exec(query, token, user.ID)
	return err
}

func (r DBUserRepository) FindByID(UserID int) (models.User, bool) {
	var user models.User
	row := r.db.QueryRow("SELECT id, token FROM users WHERE id = $1", UserID)
	err := row.Scan(&user.ID, &user.Token)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		fmt.Printf("row: %v\n", row)
		return user, false
	}
	return user, true
}

func NewDBUserRepository(db *sql.DB) (*DBUserRepository, error) {
	return &DBUserRepository{db: db}, nil
}
