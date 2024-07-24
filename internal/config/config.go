// Модуль config создает конфиг приложения
// Доступные пути конфигурации (указаны в порядке приоритета):
// 1. Поиск Переменной окружения
// 2. поиск аргумента командой строки
// 3. значение по-умолчанию

package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/caarlos0/env/v6"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/logger"
)

type argConfig struct {
	flagAAddr    string
	flagBAddr    string
	flagFAddr    string
	flagDAddr    string
	flagSAddr    bool
	flagCertAddr string
	flagKeyAddr  string
	flagCAddr    string
}

type envConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	enableHTTPS     bool   `env:"ENABLE_HTTPS"`
	keyFile         string `env:"KEY_FILE"`
	certFile        string `env:"CERT_FILE"`
	config          string `env:"CONFIG"`
}

// Config Доступные агрументы для конфигурации
type Config struct {
	// ServerAddress - Адрес запуска HTTP-сервера
	ServerAddress string `json:"server_address"`
	// BaseURL - Базовый адрес результирующего сокращённого URL
	BaseURL string `json:"base_url"`
	// FileStoragePath - Путь для сохраниния данных в файле
	FileStoragePath string `json:"file_storage_path"`
	// DatabaseDSN - Строка с адресом подключения к БД
	DatabaseDSN string `json:"database_dsn"`
	// EnableHTTPS - Включить HTTPS режим
	EnableHTTPS bool `json:"enable_https"`
	// KeyFile - путь до ключа
	KeyFile string
	// CertFile - путь до сертификата
	CertFile string
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
		flag.BoolVar(&cfg.flagSAddr, "s", false, "Включить HTTPS режим")
		flag.StringVar(&cfg.flagKeyAddr, "key", "./keyfile.pem", "Путь до ключа")
		flag.StringVar(&cfg.flagCertAddr, "cert", "./certfile.pem", "Путь до сертификата")
		flag.StringVar(&cfg.flagCAddr, "c", "", "Путь до файла конфигурации")
		flag.StringVar(&cfg.flagCAddr, "config", "", "Путь до файла конфигурации")
		// делаем разбор командной строки
		flag.Parse()
	})
	return cfg
}

// ReadConfigFromFile читает конфигурацию из файла JSON.
func ReadConfigFromFile(filePath string, config *Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}

	return nil
}

// GetConfig возвращает готовый конфиг
func GetConfig(s *logger.Logger) Config {
	argCfg := parseFlags()
	envCfg := parseEnvs(s)

	var ServerAddress string
	var BaseURL string
	var FileStoragePath string
	var DatabaseDSN string
	var enableHTTPS bool
	var keyFile string
	var certFile string
	var configPath string

	fileConfig := Config{}

	if envCfg.config != "" {
		configPath = envCfg.config
	} else if argCfg.flagCAddr != "" {
		configPath = argCfg.flagCAddr
	}

	if configPath != "" {
		err := ReadConfigFromFile(configPath, &fileConfig)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}
	}

	if fileConfig.ServerAddress != "" {
		ServerAddress = fileConfig.ServerAddress
	} else if envCfg.ServerAddress != "" {
		ServerAddress = envCfg.ServerAddress
	} else {
		ServerAddress = argCfg.flagAAddr
	}
	if fileConfig.BaseURL != "" {
		BaseURL = fileConfig.BaseURL
	} else if envCfg.BaseURL != "" {
		BaseURL = envCfg.BaseURL
	} else {
		BaseURL = argCfg.flagBAddr
	}
	if fileConfig.FileStoragePath != "" {
		FileStoragePath = fileConfig.FileStoragePath
	} else if envCfg.FileStoragePath != "" {
		FileStoragePath = envCfg.FileStoragePath
	} else {
		FileStoragePath = argCfg.flagFAddr
	}
	if fileConfig.DatabaseDSN != "" {
		DatabaseDSN = fileConfig.DatabaseDSN
	} else if envCfg.DatabaseDSN != "" {
		DatabaseDSN = envCfg.DatabaseDSN
	} else {
		DatabaseDSN = argCfg.flagDAddr
	}
	if fileConfig.EnableHTTPS {
		enableHTTPS = fileConfig.EnableHTTPS
	} else if envCfg.enableHTTPS {
		enableHTTPS = envCfg.enableHTTPS
	} else {
		enableHTTPS = argCfg.flagSAddr
	}
	if envCfg.keyFile != "" {
		keyFile = envCfg.keyFile
	} else {
		keyFile = argCfg.flagKeyAddr
	}
	if envCfg.certFile != "" {
		certFile = envCfg.certFile
	} else {
		certFile = argCfg.flagCertAddr
	}
	return Config{
		ServerAddress:   ServerAddress,
		BaseURL:         BaseURL,
		FileStoragePath: FileStoragePath,
		DatabaseDSN:     DatabaseDSN,
		EnableHTTPS:     enableHTTPS,
		KeyFile:         keyFile,
		CertFile:        certFile,
	}
}
