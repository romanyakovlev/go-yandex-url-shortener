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

func printBuildInfo() {
	formattedString := fmt.Sprintf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s",
		valueOrNA(buildVersion), valueOrNA(buildDate), valueOrNA(buildCommit),
	)
	fmt.Println(formattedString)
}

func main() {
	printBuildInfo()
	if err := server.Run(); err != nil {
		log.Fatalf("An error occurred: %v", err)
	}
}
