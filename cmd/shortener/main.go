package main

import (
	"fmt"
	"log"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func valueOrNA(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

func main() {
	formattedString := fmt.Sprintf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s",
		valueOrNA(buildVersion), valueOrNA(buildDate), valueOrNA(buildCommit),
	)
	fmt.Println(formattedString)
	if err := server.Run(); err != nil {
		log.Fatalf("An error occurred: %v", err)
	}
}
