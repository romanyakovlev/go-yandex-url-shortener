package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/server"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

func setupServer() *httptest.Server {
	sugar := logger.GetLogger()
	serverConfig := config.GetConfig(sugar)
	sharedURLRows := models.NewSharedURLRows()
	shortenerrepo := repository.MemoryURLRepository{SharedURLRows: sharedURLRows}
	userrepo := repository.MemoryUserRepository{SharedURLRows: sharedURLRows}
	shortener := service.NewURLShortenerService(serverConfig, &shortenerrepo, &userrepo)
	worker := workers.InitURLDeletionWorker(shortener)
	URLCtrl := controller.NewURLShortenerController(shortener, sugar, worker)
	db, _ := sql.Open("pgx", serverConfig.DatabaseDSN)
	defer db.Close()
	HealthCtrl := controller.NewHealthCheckController(db)
	router := server.Router(URLCtrl, HealthCtrl, sugar)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.StartDeletionWorker(ctx)
	go worker.StartErrorListener(ctx)
	return httptest.NewServer(router)
}

func Benchmark_saveURL(b *testing.B) {
	ts := setupServer()
	defer ts.Close()

	for i := 0; i < b.N; i++ {
		body := `{"url": "https://practicum.yandex.ru"}`
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			_, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

func Benchmark_getURLByID(b *testing.B) {
	ts := setupServer()
	defer ts.Close()

	body := `{"url": "https://practicum.yandex.ru"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewBufferString(body))
	resp, _ := http.DefaultClient.Do(req)
	require.NotNil(b, resp)
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	shortLinkID := strings.Split(string(respBody), "/")[len(strings.Split(string(respBody), "/"))-1]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/"+shortLinkID, nil)
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func Benchmark_shortenURL(b *testing.B) {
	ts := setupServer()
	defer ts.Close()

	for i := 0; i < b.N; i++ {
		body := `"url": "https://practicum.yandex.ru"`
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			_, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

func randomString(n int, charset string) string {
	var seededRand = rand.New(rand.NewSource(99))
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomPath(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	return "/" + randomString(length, charset)
}

func randomURL() string {
	domains := []string{"https://example.com", "https://practicum.yandex.ru", "https://google.com"}

	// Choose a random domain and path
	domain := domains[rand.Intn(len(domains))]
	path := generateRandomPath(10)

	return fmt.Sprintf("%s%s", domain, path)
}

func Benchmark_shortenBatchURL(b *testing.B) {
	ts := setupServer()
	defer ts.Close()

	rand.New(rand.NewSource(99))

	for i := 0; i < b.N; i++ {
		batchLen := 10
		batch := make([]map[string]string, batchLen)

		for j := 0; j < batchLen; j++ {
			batch[j] = map[string]string{
				"correlation_id": uuid.New().String(),
				"short_url":      randomURL(),
			}
		}

		bodyBytes, err := json.Marshal(batch)
		if err != nil {
			b.Fatalf("Failed to marshal JSON: %v", err)
		}
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewBuffer(bodyBytes))
		if err != nil {
			b.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("Failed to execute request: %v", err)
		}

		if resp != nil {
			_, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}
