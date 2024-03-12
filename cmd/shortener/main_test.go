package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/config"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/controller"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/repository"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/server"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	serverConfig := config.GetConfig()
	repo := repository.MemoryURLRepository{URLMap: make(map[string]string)}
	shortener := service.URLShortenerService{Config: serverConfig, Repo: repo}
	ctrl := controller.URLShortenerController{Shortener: shortener}
	ts = httptest.NewServer(server.Router(ctrl))

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
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, bodyIsEmpty: true},
		{method: http.MethodPost, expectedCode: http.StatusCreated, bodyIsEmpty: false},
	}

	for _, tc := range testCases {
		resp, body := testRequest(t, tc.method, "/", nil)
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
