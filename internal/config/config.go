package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type argConfig struct {
	flagAAddr string
	flagBAddr string
}

type envConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

type Config struct {
	ServerAddress string
	BaseURL       string
}

func parseEnvs() envConfig {
	var cfg envConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func parseFlags() argConfig {
	var cfg argConfig
	// указываем имя флага, значение по умолчанию и описание
	flag.StringVar(&cfg.flagAAddr, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.flagBAddr, "b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
	// делаем разбор командной строки
	flag.Parse()
	return cfg
}

func GetConfig() Config {
	argCfg := parseFlags()
	envCfg := parseEnvs()

	var ServerAddress string
	var BaseURL string

	if envCfg.ServerAddress != "" {
		ServerAddress = envCfg.ServerAddress
	} else {
		ServerAddress = argCfg.flagAAddr
	}
	if envCfg.BaseURL != "" {
		BaseURL = envCfg.BaseURL
	} else {
		BaseURL = argCfg.flagBAddr
	}
	return Config{
		ServerAddress: ServerAddress,
		BaseURL:       BaseURL,
	}
}
