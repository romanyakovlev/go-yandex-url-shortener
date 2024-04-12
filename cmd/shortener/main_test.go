package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/server"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	sugar := logger.GetLogger()
	serverConfig := config.GetConfig(sugar)
	repo := repository.MemoryURLRepository{URLMap: make(map[string]string)}
	shortener := service.URLShortenerService{Config: serverConfig, Repo: repo}
	URLCtrl := controller.URLShortenerController{Shortener: shortener, Logger: sugar}
	db, _ := sql.Open("pgx", serverConfig.DatabaseDSN)
	defer db.Close()
	HealthCtrl := controller.HealthCheckController{DB: db}
	router := server.Router(URLCtrl, HealthCtrl, sugar)
	ts = httptest.NewServer(router)

	exitCode := m.Run()

	ts.Close()

	os.Exit(exitCode)
}

func testRequest(t *testing.T, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_saveURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		bodyIsEmpty  bool
		body         string
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, bodyIsEmpty: false, body: `{"url": "https://practicum.yandex.ru"}`},
	}

	for _, tc := range testCases {
		resp, body := testRequest(t, tc.method, "/", strings.NewReader(tc.body))
		defer resp.Body.Close()

		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		if !tc.bodyIsEmpty {
			assert.NotEqual(t, body, "", "Тело ответа пустое")
		}
	}

}

func Test_getURLByID(t *testing.T) {
	linkToSave := "https://practicum.yandex.ru/"
	expectedHostName := "practicum.yandex.ru"
	resp, shortLink := testRequest(t, http.MethodPost, "/", strings.NewReader(linkToSave))
	defer resp.Body.Close()
	shortLinkID := strings.Split(shortLink, "/")[len(strings.Split(shortLink, "/"))-1]

	testCases := []struct {
		method       string
		expectedCode int
	}{
		{method: http.MethodGet, expectedCode: http.StatusOK},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPost, expectedCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range testCases {
		resp, _ := testRequest(t, tc.method, "/"+string(shortLinkID), nil)
		defer resp.Body.Close()

		if tc.method == http.MethodGet {
			assert.Equal(t, resp.Request.URL.Hostname(), expectedHostName, "Redirect не был выполнен успешно")
		}
		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
	}

}

func Test_shortenURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		bodyIsEmpty  bool
		body         string
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusInternalServerError, body: `"url": "https://practicum.yandex.ru"`, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, body: `{"url": "https://practicum.yandex.ru"}`, bodyIsEmpty: false},
	}

	for _, tc := range testCases {
		resp, body := testRequest(t, tc.method, "/api/shorten", strings.NewReader(tc.body))
		defer resp.Body.Close()

		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		if !tc.bodyIsEmpty {
			var resp models.ShortenURLResponse
			err := json.Unmarshal([]byte(body), &resp)

			assert.Equalf(t, err, nil, "Ошибка при обработке json ответа: %s", err)
			assert.NotEqual(t, resp.Result, "", "Поле 'Result' пустое")
			assert.NotEqual(t, body, "", "Тело ответа пустое")
		}
	}

}

func Test_shortenBatchURL(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		bodyIsEmpty  bool
		body         string
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusInternalServerError, body: `"url": "https://practicum.yandex.ru"`, bodyIsEmpty: true},
		{
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			body: `[{"correlation_id": "111", "short_url": "https://practicum.yandex.ru"},
					{"correlation_id": "222", "short_url": "https://yandex.ru"}]`,
			bodyIsEmpty: false,
		},
	}

	for _, tc := range testCases {
		resp, body := testRequest(t, tc.method, "/api/shorten/batch", strings.NewReader(tc.body))
		defer resp.Body.Close()

		assert.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		if !tc.bodyIsEmpty {
			var resp []models.ShortenBatchURLResponseElement
			err := json.Unmarshal([]byte(body), &resp)

			assert.Equalf(t, err, nil, "Ошибка при обработке json ответа: %s", err)
			assert.NotEqual(t, len(resp), 0, "Количество элементов в списке = 0")
			assert.NotEqual(t, body, "", "Тело ответа пустое")
		}
	}

}
