// Модуль config создает конфиг приложения
// Доступные пути конфигурации (указаны в порядке приоритета):
// 1. Поиск Переменной окружения
// 2. поиск аргумента командой строки
// 3. значение по-умолчанию

package config

import (
	"encoding/json"
	"flag"
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

type fileConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
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
	// EnableHTTPS - Включить HTTPS режим
	EnableHTTPS bool
	// KeyFile - путь до ключа
	KeyFile string
	// CertFile - путь до сертификата
	CertFile string
}

var onceParseEnvs sync.Once
var onceParseFlags sync.Once
var onceParseConfFile sync.Once

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

func parseFile(configPath string, s *logger.Logger) fileConfig {
	var cfg fileConfig
	onceParseConfFile.Do(func() {
		if configPath != "" {
			err := ReadConfigFromFile(configPath, &cfg)
			if err != nil {
				s.Fatalf("Failed to read config file: %v", err)
			}

		}
	})
	return cfg
}

// ReadConfigFromFile читает конфигурацию из файла JSON.
func ReadConfigFromFile(filePath string, config *fileConfig) error {
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

func parseFileConfig(fc *fileConfig, c *Config) {
	if fc.ServerAddress != "" {
		c.ServerAddress = fc.ServerAddress
	}
	if fc.BaseURL != "" {
		c.BaseURL = fc.BaseURL
	}
	if fc.FileStoragePath != "" {
		c.FileStoragePath = fc.FileStoragePath
	}
	if fc.DatabaseDSN != "" {
		c.DatabaseDSN = fc.DatabaseDSN
	}
	if fc.EnableHTTPS {
		c.EnableHTTPS = fc.EnableHTTPS
	}
}

func parseEnvConfig(ec *envConfig, c *Config) {
	if ec.ServerAddress != "" {
		c.ServerAddress = ec.ServerAddress
	}
	if ec.BaseURL != "" {
		c.BaseURL = ec.BaseURL
	}
	if ec.FileStoragePath != "" {
		c.FileStoragePath = ec.FileStoragePath
	}
	if ec.DatabaseDSN != "" {
		c.DatabaseDSN = ec.DatabaseDSN
	}
	if ec.enableHTTPS {
		c.EnableHTTPS = ec.enableHTTPS
	}
	if ec.keyFile != "" {
		c.KeyFile = ec.keyFile
	}
	if ec.certFile != "" {
		c.CertFile = ec.certFile
	}
}

func parseArgConfig(ac *argConfig, c *Config) {
	if ac.flagAAddr != "" {
		c.ServerAddress = ac.flagAAddr
	}
	if ac.flagBAddr != "" {
		c.BaseURL = ac.flagBAddr
	}
	if ac.flagFAddr != "" {
		c.FileStoragePath = ac.flagFAddr
	}
	if ac.flagDAddr != "" {
		c.DatabaseDSN = ac.flagDAddr
	}
	if ac.flagSAddr {
		c.EnableHTTPS = ac.flagSAddr
	}
	if ac.flagKeyAddr != "" {
		c.KeyFile = ac.flagKeyAddr
	}
	if ac.flagCertAddr != "" {
		c.CertFile = ac.flagCertAddr
	}
}

// GetConfig возвращает готовый конфиг
func GetConfig(s *logger.Logger) Config {
	var configPath string
	config := Config{}
	argCfg := parseFlags()
	envCfg := parseEnvs(s)
	if envCfg.config != "" {
		configPath = envCfg.config
	} else if argCfg.flagCAddr != "" {
		configPath = argCfg.flagCAddr
	}
	fileCfg := parseFile(configPath, s)
	parseFileConfig(&fileCfg, &config)
	parseArgConfig(&argCfg, &config)
	parseEnvConfig(&envCfg, &config)
	return config
}
