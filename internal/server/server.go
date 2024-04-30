package server

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/db"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	shortenerRepository "github.com/romanyakovlev/go-yandex-url-shortener/internal/repository/shortener"
	userRepository "github.com/romanyakovlev/go-yandex-url-shortener/internal/repository/user"
	shortenerService "github.com/romanyakovlev/go-yandex-url-shortener/internal/service/shortener"
	userService "github.com/romanyakovlev/go-yandex-url-shortener/internal/service/user"
)

func Router(
	URLShortenerController *controller.URLShortenerController,
	HealthCheckController *controller.HealthCheckController,
	userService *userService.UserService,
	sugar *logger.Logger,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLoggerMiddleware(sugar))
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.JWTMiddleware(userService))
	r.Post("/", URLShortenerController.SaveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", URLShortenerController.GetURLByID)
	r.Post("/api/shorten/batch", URLShortenerController.ShortenBatchURL)
	r.Post("/api/shorten", URLShortenerController.ShortenURL)
	r.Get("/api/user/urls", URLShortenerController.GetURLByUser)
	r.Get("/ping", HealthCheckController.Ping)
	return r
}

func initURLRepository(serverConfig config.Config, db *sql.DB, sugar *logger.Logger) (shortenerService.URLRepository, error) {
	if serverConfig.DatabaseDSN != "" {
		return shortenerRepository.NewDBURLRepository(db)
	} else if serverConfig.FileStoragePath != "" {
		return shortenerRepository.NewFileURLRepository(serverConfig, sugar)
	} else {
		return shortenerRepository.NewMemoryURLRepository()
	}

}

func initUserRepository(serverConfig config.Config, db *sql.DB, sugar *logger.Logger) (userService.UserRepository, error) {
	if serverConfig.DatabaseDSN != "" {
		return userRepository.NewDBUserRepository(db)
		//} else if serverConfig.FileStoragePath != "" {
		//	return userRepository.NewFileUserRepository(serverConfig, sugar)
	} else {
		return userRepository.NewMemoryUserRepository()
	}
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

	shortenerrepo, err := initURLRepository(serverConfig, DB, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	userrepo, err := initUserRepository(serverConfig, DB, sugar)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
		return err
	}
	shortenerservice := shortenerService.NewURLShortenerService(serverConfig, shortenerrepo)
	userservice := userService.NewUserService(serverConfig, userrepo)
	URLCtrl := controller.NewURLShortenerController(shortenerservice, sugar)
	HealthCtrl := controller.NewHealthCheckController(DB)
	router := Router(URLCtrl, HealthCtrl, userservice, sugar)
	err = http.ListenAndServe(serverConfig.ServerAddress, router)
	if err != nil {
		sugar.Errorf("Server error: %v", err)
	}
	return err
}
