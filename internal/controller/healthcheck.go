package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// HealthCheckController Контроллер для проверки «состояния здоровья» системы
type HealthCheckController struct {
	db *sql.DB
}

// Ping проверяет:
// 1. подключение к БД
func (hc HealthCheckController) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := hc.db.PingContext(ctx); err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// NewHealthCheckController создает HealthCheckController
func NewHealthCheckController(db *sql.DB) *HealthCheckController {
	return &HealthCheckController{db: db}
}
