package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
	"go.uber.org/zap"
	"net/http"
)

func Router(controller controller.URLShortenerController, sugar *zap.SugaredLogger) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLoggerMiddleware(sugar))
	r.Use(middlewares.GzipMiddleware)
	r.Post("/", controller.SaveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", controller.GetURLByID)
	r.Post("/api/shorten", controller.ShortenURL)
	return r
}

func Run() error {
	sugar := logger.GetLogger()
	serverConfig := config.GetConfig(sugar)
	repo := repository.MemoryURLRepository{URLMap: make(map[string]string)}
	shortener := service.URLShortenerService{Config: serverConfig, Repo: repo}
	ctrl := controller.URLShortenerController{Shortener: shortener, Logger: sugar}
	router := Router(ctrl, sugar)
	err := http.ListenAndServe(serverConfig.ServerAddress, router)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
	}
	return err
}
