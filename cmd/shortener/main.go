package main

import (
	"log"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("An error occurred: %v", err)
	}
}
