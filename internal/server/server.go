package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/db"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

func Router(
	URLShortenerController controller.URLShortenerController,
	HealthCheckController controller.HealthCheckController,
	sugar *logger.Logger,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLoggerMiddleware(sugar))
	r.Use(middlewares.GzipMiddleware)
	r.Post("/", URLShortenerController.SaveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", URLShortenerController.GetURLByID)
	r.Post("/api/shorten", URLShortenerController.ShortenURL)
	r.Get("/ping", HealthCheckController.Ping)
	return r
}

func Run() error {
	sugar := logger.GetLogger()
	serverConfig := config.GetConfig(sugar)

	DB, err := db.InitDB(serverConfig.DatabaseDSN, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	
	defer DB.Close()

	migrationsDir := "./migrations"

	if err := goose.Up(DB, migrationsDir); err != nil {
		sugar.Fatalf("goose Up failed: %v\n", err)
	}

	repo, err := repository.NewURLRepository(serverConfig, DB, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	shortener := service.URLShortenerService{Config: serverConfig, Repo: repo}
	URLCtrl := controller.URLShortenerController{Shortener: shortener, Logger: sugar}
	HealthCtrl := controller.HealthCheckController{DB: DB}
	router := Router(URLCtrl, HealthCtrl, sugar)
	err = http.ListenAndServe(serverConfig.ServerAddress, router)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
	}
	return err
}
