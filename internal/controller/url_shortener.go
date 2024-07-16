// Package controller пакет с хэндлерами http-запросов
package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/apperrors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

// URLShortener Интерфейс сервиса сокращения ссылок
type URLShortener interface {
	// AddURL добавление url
	AddURL(urlStr string) (models.SavedURL, error)
	// AddBatchURL добавление списка url
	AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.CorrelationSavedURL, error)
	// AddUserToURL присвоение url пользователю
	AddUserToURL(SavedURL models.SavedURL, user models.User) error
	// AddBatchUserToURL присвоение списка url пользователю
	AddBatchUserToURL(SavedURLs []models.SavedURL, user models.User) error
	// GetURL Получение url по короткой ссылке
	GetURL(shortURL string) (models.URLRow, bool)
	// GetURLByUser Получение всех url, присвоенных пользователю
	GetURLByUser(user models.User) ([]models.URLByUserResponseElement, bool)
	// GetURLByOriginalURL Получение короткой ссылки для url
	GetURLByOriginalURL(originalURL string) (string, bool)
	// DeleteBatchURL удаление списка url
	DeleteBatchURL(urls []string, user models.User) error
	// ConvertCorrelationSavedURLsToResponse преобразование модели данных []models.CorrelationSavedURL
	// в response-модель []models.ShortenBatchURLResponseElement для API-хелдлера
	ConvertCorrelationSavedURLsToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement
	// ConvertCorrelationSavedURLsToSavedURL преобразование модели данных []models.CorrelationSavedURL
	// в модель []models.SavedURL для API-хелдлера
	ConvertCorrelationSavedURLsToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL
}

// URLShortenerController Контроллер для взаимодействия с внутренним сервисом сокращения ссылок URLShortener
type URLShortenerController struct {
	shortener URLShortener
	worker    *workers.URLDeletionWorker
	logger    *logger.Logger
}

// SaveURL Принимает url и возвращает короткую ссылку (ожидает url в text/plain body)
func (c URLShortenerController) SaveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	savedURL, err := c.shortener.AddURL(urlStr)
	if err != nil {
		var appError *apperrors.OriginalURLAlreadyExists
		if ok := errors.As(err, &appError); ok {
			c.logger.Debugf("Shortener service error: %s", err)
			value, ok := c.shortener.GetURLByOriginalURL(urlStr)
			if ok {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprintf(w, "%v", value)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			c.logger.Debugf("Shortener service error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", savedURL.ShortURL)
	user, _ := middlewares.GetUserFromContext(r.Context())
	if err := c.shortener.AddUserToURL(savedURL, user); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "something went wrong: %s", nil)
	}
}

// DeleteBatchURL Удаляет список url
func (c URLShortenerController) DeleteBatchURL(w http.ResponseWriter, r *http.Request) {
	var urls []string
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "cannot decode request JSON body: %s", nil)
		return
	}
	user, _ := middlewares.GetUserFromContext(r.Context())
	req := workers.DeletionRequest{User: user, URLs: urls}
	if err := c.worker.SendDeletionRequestToWorker(req); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "error sending to deletion worker request: %s", nil)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

// GetURLByID возвращает url на основе короткой ссылки
func (c URLShortenerController) GetURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	urlRow, ok := c.shortener.GetURL(shortURL)
	if urlRow.DeletedFlag {
		w.WriteHeader(http.StatusGone)
		return
	}
	if ok {
		w.Header().Set("Location", urlRow.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// GetURLByUser возвращает список url, которые пользователь загрузил в систему
func (c URLShortenerController) GetURLByUser(w http.ResponseWriter, r *http.Request) {
	user, _ := middlewares.GetUserFromContext(r.Context())
	resp, ok := c.shortener.GetURLByUser(user)
	if ok {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			c.handleError(w, err, http.StatusInternalServerError, "cannot encode response JSON body: %s", nil)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// ShortenURL Принимает url и возвращает короткую ссылку (ожидает url в json body)
func (c URLShortenerController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "cannot decode request JSON body: %s", nil)
		return
	}
	savedURL, err := c.shortener.AddURL(req.URL)
	if err != nil {
		var appError *apperrors.OriginalURLAlreadyExists
		if ok := errors.As(err, &appError); ok {
			c.logger.Debugf("Shortener service error: %s", err)
			value, ok := c.shortener.GetURLByOriginalURL(req.URL)
			if ok {
				resp := models.ShortenURLResponse{Result: value}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				enc := json.NewEncoder(w)
				if err := enc.Encode(resp); err != nil {
					c.logger.Debugf("cannot encode response JSON body: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			c.logger.Debugf("Shortener service error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	resp := models.ShortenURLResponse{Result: savedURL.ShortURL}
	user, _ := middlewares.GetUserFromContext(r.Context())
	if err := c.shortener.AddUserToURL(savedURL, user); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "something went wrong: %s", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "cannot encode response JSON body: %s", nil)
	}
}

// ShortenBatchURL Принимает список url в формате json и возвращает список коротких ссылок
func (c URLShortenerController) ShortenBatchURL(w http.ResponseWriter, r *http.Request) {
	var req []models.ShortenBatchURLRequestElement
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "cannot decode request JSON body: %s", nil)
		return
	}
	user, _ := middlewares.GetUserFromContext(r.Context())
	correlationSavedURLs, err := c.shortener.AddBatchURL(req)
	resp := c.shortener.ConvertCorrelationSavedURLsToResponse(correlationSavedURLs)
	if err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "Shortener service error: %s", nil)
		return
	}
	savedURLs := c.shortener.ConvertCorrelationSavedURLsToSavedURL(correlationSavedURLs)
	if err := c.shortener.AddBatchUserToURL(savedURLs, user); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "something went wrong: %s", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.handleError(w, err, http.StatusInternalServerError, "cannot encode response JSON body: %s", nil)
	}
}

func (c URLShortenerController) handleError(w http.ResponseWriter, err error, statusCode int, logMessage string, resp interface{}) {
	c.logger.Debugf(logMessage, err)
	w.WriteHeader(statusCode)
	if resp != nil {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if encodeErr := enc.Encode(resp); encodeErr != nil {
			c.logger.Debugf("cannot encode response JSON body: %s", encodeErr)
		}
	}
}

// NewURLShortenerController создает URLShortenerController
func NewURLShortenerController(shortener URLShortener, logger *logger.Logger, worker *workers.URLDeletionWorker) *URLShortenerController {
	return &URLShortenerController{shortener: shortener, logger: logger, worker: worker}
}
