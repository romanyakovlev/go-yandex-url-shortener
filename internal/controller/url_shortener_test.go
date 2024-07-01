package controller
/*
import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/apperrors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

// MockURLShortener is a mock implementation of URLShortener interface
type MockURLShortener struct {
	mock.Mock
}

func (m *MockURLShortener) AddURL(urlStr string) (models.SavedURL, error) {
	args := m.Called(urlStr)
	return args.Get(0).(models.SavedURL), args.Error(1)
}

func (m *MockURLShortener) AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.CorrelationSavedURL, error) {
	args := m.Called(batchArray)
	return args.Get(0).([]models.CorrelationSavedURL), args.Error(1)
}

func (m *MockURLShortener) AddUserToURL(SavedURL models.SavedURL, user models.User) error {
	args := m.Called(SavedURL, user)
	return args.Error(0)
}

func (m *MockURLShortener) AddBatchUserToURL(SavedURLs []models.SavedURL, user models.User) error {
	args := m.Called(SavedURLs, user)
	return args.Error(0)
}

func (m *MockURLShortener) GetURL(shortURL string) (models.URLRow, bool) {
	args := m.Called(shortURL)
	return args.Get(0).(models.URLRow), args.Bool(1)
}

func (m *MockURLShortener) GetURLByUser(user models.User) ([]models.URLByUserResponseElement, bool) {
	args := m.Called(user)
	return args.Get(0).([]models.URLByUserResponseElement), args.Bool(1)
}

func (m *MockURLShortener) GetURLByOriginalURL(originalURL string) (string, bool) {
	args := m.Called(originalURL)
	return args.String(0), args.Bool(1)
}

func (m *MockURLShortener) DeleteBatchURL(urls []string, user models.User) error {
	args := m.Called(urls, user)
	return args.Error(0)
}

func (m *MockURLShortener) ConvertCorrelationSavedURLToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement {
	args := m.Called(correlationSavedURLs)
	return args.Get(0).([]models.ShortenBatchURLResponseElement)
}

func (m *MockURLShortener) ConvertCorrelationSavedURLToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL {
	args := m.Called(correlationSavedURLs)
	return args.Get(0).([]models.SavedURL)
}

// TestSaveURL tests the SaveURL method of URLShortenerController
func TestSaveURL(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(workers.URLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	url := "http://example.com"
	user := models.User{ID: "testuser"}
	expectedSavedURL := models.SavedURL{ShortURL: "http://short.url"}
	mockShortener.On("AddURL", url).Return(expectedSavedURL, nil)
	mockShortener.On("AddUserToURL", expectedSavedURL, user).Return(nil)

	reqBody, _ := json.Marshal(map[string]string{"url": url})
	req, err := http.NewRequest("POST", "/save-url", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	ctx := mock.AnythingOfType("*context.emptyCtx")
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.SaveURL)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, expectedSavedURL.ShortURL, recorder.Body.String())

	mockShortener.AssertExpectations(t)
}

// TestDeleteBatchURL tests the DeleteBatchURL method of URLShortenerController
func TestDeleteBatchURL(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(MockURLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	urls := []string{"http://short.url/1", "http://short.url/2"}
	user := models.User{ID: "testuser"}
	mockWorker.On("SendDeletionRequestToWorker", mock.Anything).Return(nil)

	reqBody, _ := json.Marshal(urls)
	req, err := http.NewRequest("DELETE", "/delete-batch", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	ctx := mock.AnythingOfType("*context.emptyCtx")
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.DeleteBatchURL)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusAccepted, recorder.Code)

	mockWorker.AssertExpectations(t)
}

// TestGetURLByID tests the GetURLByID method of URLShortenerController
func TestGetURLByID(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(MockURLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	expectedURLRow := models.URLRow{OriginalURL: "http://example.com", DeletedFlag: false}
	mockShortener.On("GetURL", "shortURL").Return(expectedURLRow, true)

	req, err := http.NewRequest("GET", "/get-url/shortURL", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.GetURLByID)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, expectedURLRow.OriginalURL, recorder.Header().Get("Location"))

	mockShortener.AssertExpectations(t)
}

// TestGetURLByUser tests the GetURLByUser method of URLShortenerController
func TestGetURLByUser(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(MockURLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	expectedResponse := []models.URLByUserResponseElement{{ShortURL: "http://short.url/1"}, {ShortURL: "http://short.url/2"}}
	mockShortener.On("GetURLByUser", mock.Anything).Return(expectedResponse, true)

	req, err := http.NewRequest("GET", "/get-url-by-user", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := mock.AnythingOfType("*context.emptyCtx")
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.GetURLByUser)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	mockShortener.AssertExpectations(t)
}

// TestShortenURL tests the ShortenURL method of URLShortenerController
func TestShortenURL(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(MockURLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	url := "http://example.com"
	user := models.User{ID: "testuser"}
	expectedSavedURL := models.SavedURL{ShortURL: "http://short.url"}
	mockShortener.On("AddURL", url).Return(expectedSavedURL, nil)
	mockShortener.On("AddUserToURL", expectedSavedURL, user).Return(nil)

	reqBody, _ := json.Marshal(map[string]string{"url": url})
	req, err := http.NewRequest("POST", "/shorten-url", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	ctx := mock.AnythingOfType("*context.emptyCtx")
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.ShortenURL)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.JSONEq(t, `{"result":"http://short.url"}`, recorder.Body.String())

	mockShortener.AssertExpectations(t)
}

// TestShortenBatchURL tests the ShortenBatchURL method of URLShortenerController
func TestShortenBatchURL(t *testing.T) {
	mockShortener := new(MockURLShortener)
	mockLogger := logger.NewFakeLogger()
	mockWorker := new(MockURLDeletionWorker)
	controller := NewURLShortenerController(mockShortener, mockLogger, mockWorker)

	request := []models.ShortenBatchURLRequestElement{{URL: "http://example.com/1"}, {URL: "http://example.com/2"}}
	user := models.User{ID: "testuser"}
	expectedCorrelationSavedURLs := []models.CorrelationSavedURL{
		{OriginalURL: "http://example.com/1", ShortURL: "http://short.url/1"},
		{OriginalURL: "http://example.com/2", ShortURL: "http://short.url/2"},
	}
	mockShortener.On("AddBatchURL", request).Return(expectedCorrelationSavedURLs, nil)
	mockShortener.On("ConvertCorrelationSavedURLToResponse", expectedCorrelationSavedURLs).
		Return([]models.ShortenBatchURLResponseElement{
			{OriginalURL: "http://example.com/1", ShortURL: "http://short.url/1"},
			{OriginalURL: "http://example.com/2", ShortURL: "http://short.url/2"},
		})
	mockShortener.On("ConvertCorrelationSavedURLToSavedURL", expectedCorrelationSavedURLs).
		Return([]models.SavedURL{
			{ShortURL: "http://short.url/1"},
			{ShortURL: "http://short.url/2"},
		})
	mockShortener.On("AddBatchUserToURL", mock.Anything, user).Return(nil)

	reqBody, _ := json.Marshal(request)
	req, err := http.NewRequest("POST", "/shorten-batch-url", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	ctx := mock.AnythingOfType("*context.emptyCtx")
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.ShortenBatchURL)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	mockShortener.AssertExpectations(t)
}

func TestMain(m *testing.M) {
	m.Run()
}
*/
