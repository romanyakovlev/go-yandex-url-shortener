package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/db"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

func Example() {
	serverConfig := config.Config{
		ServerAddress: "localhost:8080", // Адрес сервера
		DatabaseDSN:   "mock_db_dsn",    // DSN базы данных
	}

	sugar := logger.GetLogger() // Получаем логгер

	DB, err := db.InitDB(serverConfig.DatabaseDSN, sugar) // Инициализация базы данных
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	defer DB.Close() // Закрываем базу данных по завершении

	sharedURLRows := models.NewSharedURLRows() // Создаем объект sharedURLRows

	shortenerrepo := repository.MemoryURLRepository{SharedURLRows: sharedURLRows} // Репозиторий URL
	userrepo := repository.MemoryUserRepository{SharedURLRows: sharedURLRows}     // Репозиторий пользователей

	// Инициализируем сервис URL-сокращателя
	shortenerService := service.NewURLShortenerService(serverConfig, &shortenerrepo, &userrepo)

	// Инициализируем рабочего для удаления URL
	worker := workers.InitURLDeletionWorker(shortenerService)

	// Инициализируем контроллер URL
	URLCtrl := controller.NewURLShortenerController(shortenerService, sugar, worker)

	// Инициализируем контроллер проверки состояния здоровья
	HealthCtrl := controller.NewHealthCheckController(DB)

	// Инициализируем маршрутизатор
	router := Router(URLCtrl, HealthCtrl, sugar, &serverConfig)

	// Контекст для грациозного завершения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем рабочего для удаления URL
	go worker.StartDeletionWorker(ctx)

	// Запускаем прослушиватель ошибок рабочего
	go worker.StartErrorListener(ctx)

	// Запускаем HTTP сервер
	go func() {
		err := http.ListenAndServe(serverConfig.ServerAddress, router)
		if err != nil {
			log.Fatalf("Server error: %v", err) // Ошибка сервера
		}
	}()

	// Ждем запуска сервера (в реальном примере можно использовать более надежный механизм синхронизации)
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Server started successfully") // Сервер успешно запущен
}
