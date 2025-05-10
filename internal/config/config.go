package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config содержит настройки приложения
type Config struct {
	// Пути к директориям
	IncomingDir string `json:"incoming_dir"`
	OutgoingDir string `json:"outgoing_dir"`
	ArchiveDir  string `json:"archive_dir"`
	LogsDir     string `json:"logs_dir"`

	// Настройки базы данных
	DatabasePath string `json:"database_path"`

	// Настройки веб-сервера
	ServerPort string `json:"server_port"`

	// Настройки сканера
	ScannerAddress string `json:"scanner_address"`
	ScanCommand    string `json:"scan_command"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	// Базовые пути относительно текущей директории
	baseDir := "data"

	return Config{
		IncomingDir:    filepath.Join(baseDir, "incoming"),
		OutgoingDir:    filepath.Join(baseDir, "outgoing"),
		ArchiveDir:     filepath.Join(baseDir, "archive"),
		LogsDir:        filepath.Join(baseDir, "logs"),
		DatabasePath:   filepath.Join(baseDir, "filemarker.db"),
		ServerPort:     "8080",
		ScannerAddress: "127.0.0.1:2001",
		ScanCommand:    " ", // Пробел как простейшая команда сканирования
	}
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(path string) (Config, error) {
	// Если файл не найден, используем конфигурацию по умолчанию
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultConfig := DefaultConfig()

		// Создаем директории, если не существуют
		ensureDirectoryExists(defaultConfig.IncomingDir)
		ensureDirectoryExists(defaultConfig.OutgoingDir)
		ensureDirectoryExists(defaultConfig.ArchiveDir)
		ensureDirectoryExists(defaultConfig.LogsDir)

		// Создаем файл с конфигурацией по умолчанию
		if err := SaveConfig(path, defaultConfig); err != nil {
			return defaultConfig, err
		}

		return defaultConfig, nil
	}

	// Читаем файл
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), err
	}

	// Разбираем JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	// Создаем директории, если не существуют
	ensureDirectoryExists(config.IncomingDir)
	ensureDirectoryExists(config.OutgoingDir)
	ensureDirectoryExists(config.ArchiveDir)
	ensureDirectoryExists(config.LogsDir)

	return config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(path string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ensureDirectoryExists создает директорию, если она не существует
func ensureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}
