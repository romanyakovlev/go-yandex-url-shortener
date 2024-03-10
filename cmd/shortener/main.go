package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urlMap = map[string]string{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func saveURL(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	urlStr := string(bytes)
	randomPath := RandStringBytes(8)
	urlMap[randomPath] = urlStr
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%v", randomPath)
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")
	value, ok := urlMap[shortURL]
	if ok {
		w.Header().Set("Location", value)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func URLShortenerRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", saveURL)
	r.Get("/{shortURL:[A-Za-z]{8}}", getURLByID)
	return r
}

func main() {
	err := http.ListenAndServe(`:8080`, URLShortenerRouter())
	if err != nil {
		panic(err)
	}
}
