// Модуль config создает конфиг приложения
// Доступные пути конфигурации (указаны в порядке приоритета):
// 1. Поиск Переменной окружения
// 2. поиск аргумента командой строки
// 3. значение по-умолчанию

package config

import (
	"flag"
	"sync"

	"github.com/caarlos0/env/v6"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

type argConfig struct {
	flagAAddr string
	flagBAddr string
	flagFAddr string
	flagDAddr string
}

type envConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

// Config Доступные агрументы для конфигурации
type Config struct {
	// ServerAddress - Адрес запуска HTTP-сервера
	ServerAddress string
	// BaseURL - Базовый адрес результирующего сокращённого URL
	BaseURL string
	// FileStoragePath - Путь для сохраниния данных в файле
	FileStoragePath string
	// DatabaseDSN - Строка с адресом подключения к БД
	DatabaseDSN string
}

var onceParseEnvs sync.Once
var onceParseFlags sync.Once

func parseEnvs(s *logger.Logger) envConfig {
	var cfg envConfig
	onceParseEnvs.Do(func() {
		err := env.Parse(&cfg)
		if err != nil {
			s.Fatal(err)
		}
	})
	return cfg
}

func parseFlags() argConfig {
	var cfg argConfig
	onceParseFlags.Do(func() {
		// указываем имя флага, значение по умолчанию и описание
		flag.StringVar(&cfg.flagAAddr, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
		flag.StringVar(&cfg.flagBAddr, "b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
		flag.StringVar(&cfg.flagFAddr, "f", "", "Путь для сохраниния данных в файле")
		flag.StringVar(&cfg.flagDAddr, "d", "", "Строка с адресом подключения к БД")
		// делаем разбор командной строки
		flag.Parse()
	})
	return cfg
}

func GetConfig(s *logger.Logger) Config {
	argCfg := parseFlags()
	envCfg := parseEnvs(s)

	var ServerAddress string
	var BaseURL string
	var FileStoragePath string
	var DatabaseDSN string

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
	if envCfg.FileStoragePath != "" {
		FileStoragePath = envCfg.FileStoragePath
	} else {
		FileStoragePath = argCfg.flagFAddr
	}
	if envCfg.DatabaseDSN != "" {
		DatabaseDSN = envCfg.DatabaseDSN
	} else {
		DatabaseDSN = argCfg.flagDAddr
	}
	return Config{
		ServerAddress:   ServerAddress,
		BaseURL:         BaseURL,
		FileStoragePath: FileStoragePath,
		DatabaseDSN:     DatabaseDSN,
	}
}
