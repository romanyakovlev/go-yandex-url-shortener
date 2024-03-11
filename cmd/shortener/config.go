package main

import (
	"flag"
)

var flagAAddr string
var flagBAddr string

func parseFlags() {
	// указываем имя флага, значение по умолчанию и описание
	flag.StringVar(&flagAAddr, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&flagBAddr, "b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
	// делаем разбор командной строки
	flag.Parse()
}
