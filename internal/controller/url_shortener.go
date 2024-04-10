package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/service"
)

type URLShortenerController struct {
	Shortener service.URLShortenerService
	Logger    *logger.Logger
}

func (c URLShortenerController) SaveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	shortURL, err := c.Shortener.AddURL(urlStr)
	if err != nil {
		c.Logger.Debugf("Shortener service error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", shortURL)
}

func (c URLShortenerController) GetURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	value, ok := c.Shortener.GetURL(shortURL)
	if ok {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (c URLShortenerController) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenURLRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		c.Logger.Debugf("cannot decode request JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	shortURL, err := c.Shortener.AddURL(req.URL)
	if err != nil {
		c.Logger.Debugf("Shortener service error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := models.ShortenURLResponse{Result: shortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		c.Logger.Debugf("cannot encode response JSON body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
