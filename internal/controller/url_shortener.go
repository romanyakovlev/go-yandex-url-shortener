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

type URLShortener interface {
	AddURL(urlStr string) (models.SavedURL, error)
	AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.CorrelationSavedURL, error)
	AddUserToURL(SavedURL models.SavedURL, user models.User) error
	AddBatchUserToURL(SavedURLs []models.SavedURL, user models.User) error
	GetURL(shortURL string) (models.URLRow, bool)
	GetURLByUser(user models.User) ([]models.URLByUserResponseElement, bool)
	GetURLByOriginalURL(originalURL string) (string, bool)
	DeleteBatchURL(urls []string, user models.User) error
	ConvertCorrelationSavedURLToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement
	ConvertCorrelationSavedURLToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL
}

type URLShortenerController struct {
	shortener URLShortener
	worker    *workers.URLDeletionWorker
	logger    *logger.Logger
}

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
	err = c.shortener.AddUserToURL(savedURL, user)
	if err != nil {
		c.logger.Debugf("something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c URLShortenerController) DeleteBatchURL(w http.ResponseWriter, r *http.Request) {
	var urls []string
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&urls); err != nil {
		c.logger.Debugf("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, _ := middlewares.GetUserFromContext(r.Context())
	req := workers.DeletionRequest{
		User: user,
		URLs: urls,
	}
	if err := c.worker.SendDeletionRequestToWorker(req); err != nil {
		c.logger.Debugf("error send to deletion worker request: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

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

func (c URLShortenerController) GetURLByUser(w http.ResponseWriter, r *http.Request) {
	user, _ := middlewares.GetUserFromContext(r.Context())
	resp, ok := c.shortener.GetURLByUser(user)
	if ok {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			c.logger.Debugf("cannot encode response JSON body: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (c URLShortenerController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenURLRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		c.logger.Debugf("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	w.Header().Set("Content-Type", "application/json")
	user, _ := middlewares.GetUserFromContext(r.Context())
	err = c.shortener.AddUserToURL(savedURL, user)
	if err != nil {
		c.logger.Debugf("something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		c.logger.Debugf("cannot encode response JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c URLShortenerController) ShortenBatchURL(w http.ResponseWriter, r *http.Request) {
	var req []models.ShortenBatchURLRequestElement
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		c.logger.Debugf("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, _ := middlewares.GetUserFromContext(r.Context())
	correlationSavedURLs, err := c.shortener.AddBatchURL(req)
	resp := c.shortener.ConvertCorrelationSavedURLToResponse(correlationSavedURLs)
	if err != nil {
		c.logger.Debugf("Shortener service error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	savedURLs := c.shortener.ConvertCorrelationSavedURLToSavedURL(correlationSavedURLs)
	err = c.shortener.AddBatchUserToURL(savedURLs, user)
	if err != nil {
		c.logger.Debugf("something went wrong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		c.logger.Debugf("cannot encode response JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func NewURLShortenerController(shortener URLShortener, logger *logger.Logger, worker *workers.URLDeletionWorker) *URLShortenerController {
	return &URLShortenerController{shortener: shortener, logger: logger, worker: worker}
}
