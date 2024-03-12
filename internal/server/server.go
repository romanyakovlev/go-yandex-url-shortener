package server

import (
	"log"
	"net/http"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

func Router(controller controller.URLShortenerController) chi.Router {
	r := chi.NewRouter()
	r.Post("/", controller.SaveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", controller.GetURLByID)
	return r
}

func Run() error {
	serverConfig := config.GetConfig()
	repo := repository.MemoryURLRepository{URLMap: make(map[string]string)}
	shortener := service.URLShortenerService{Config: serverConfig, Repo: repo}
	ctrl := controller.URLShortenerController{Shortener: shortener}
	err := http.ListenAndServe(serverConfig.ServerAddress, Router(ctrl))
	if err != nil {
		log.Printf("Server error: %v", err)
	}
	return err
}
