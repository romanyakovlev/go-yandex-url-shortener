// Package server выполняет инициализацию web-приложения
package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/db"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

// Router настраивает и возвращает маршрутизатор для web-приложения.
func Router(
	URLShortenerController *controller.URLShortenerController,
	HealthCheckController *controller.HealthCheckController,
	sugar *logger.Logger,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLoggerMiddleware(sugar))
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.JWTMiddleware)
	r.Post("/", URLShortenerController.SaveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", URLShortenerController.GetURLByID)
	r.Post("/api/shorten/batch", URLShortenerController.ShortenBatchURL)
	r.Post("/api/shorten", URLShortenerController.ShortenURL)
	r.Get("/api/user/urls", URLShortenerController.GetURLByUser)
	r.Delete("/api/user/urls", URLShortenerController.DeleteBatchURL)
	r.Get("/ping", HealthCheckController.Ping)
	return r
}

// InitURLRepository инициализирует репозиторий URL в зависимости от конфигурации.
func InitURLRepository(serverConfig config.Config, db *sql.DB, sharedURLRows *models.SharedURLRows, sugar *logger.Logger) (service.URLRepository, error) {
	if serverConfig.DatabaseDSN != "" {
		return repository.NewDBURLRepository(db)
	} else if serverConfig.FileStoragePath != "" {
		return repository.NewFileURLRepository(serverConfig, sugar)
	} else {
		return repository.NewMemoryURLRepository(sharedURLRows)
	}

}

// InitURLRepository инициализирует репозиторий пользователя в зависимости от конфигурации.
func initUserRepository(serverConfig config.Config, db *sql.DB, sharedURLRows *models.SharedURLRows, sugar *logger.Logger) (service.UserRepository, error) {
	if serverConfig.DatabaseDSN != "" {
		return repository.NewDBUserRepository(db)
	} else if serverConfig.FileStoragePath != "" {
		return repository.NewFileUserRepository(serverConfig, sugar)
	} else {
		return repository.NewMemoryUserRepository(sharedURLRows)
	}

}

// Run запускает web-приложение.
func Run() error {
	sugar := logger.GetLogger()
	serverConfig := config.GetConfig(sugar)

	DB, err := db.InitDB(serverConfig.DatabaseDSN, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	defer DB.Close()
	sharedURLRows := models.NewSharedURLRows()

	shortenerrepo, err := InitURLRepository(serverConfig, DB, sharedURLRows, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	userrepo, err := initUserRepository(serverConfig, DB, sharedURLRows, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	shortenerService := service.NewURLShortenerService(serverConfig, shortenerrepo, userrepo)
	worker := workers.InitURLDeletionWorker(shortenerService)
	URLCtrl := controller.NewURLShortenerController(shortenerService, sugar, worker)
	HealthCtrl := controller.NewHealthCheckController(DB)
	router := Router(URLCtrl, HealthCtrl, sugar)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.StartDeletionWorker(ctx)
	go worker.StartErrorListener(ctx)
	server := &http.Server{
		Addr:    serverConfig.ServerAddress,
		Handler: router,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	go func() {
		if serverConfig.EnableHTTPS {
			if serverConfig.CertFile == "" || serverConfig.KeyFile == "" {
				log.Fatal("certFile and keyFile must be provided when HTTPS mode is enabled")
			}
			log.Println("HTTPS mode is enabled")
			if err := server.ListenAndServeTLS(serverConfig.CertFile, serverConfig.KeyFile); err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			log.Println("HTTPS mode is not enabled")
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
		return err
	}
	log.Println("Server exited properly")
	return nil
}
