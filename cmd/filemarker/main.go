package main

import (
	"FileMarker/internal/config"
	"FileMarker/internal/database"
	"FileMarker/internal/web"
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Парсим флаги командной строки
	configPath := flag.String("config", "config.json", "Путь к файлу конфигурации")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации: %v. Используем значения по умолчанию.", err)
		cfg = config.DefaultConfig()
	}

	// Настраиваем логгирование
	logFile, err := setupLogging(cfg.LogsDir)
	if err != nil {
		log.Printf("Ошибка настройки логгирования: %v. Логи будут выводиться только в консоль.", err)
	} else {
		defer logFile.Close()
	}

	// Инициализируем базу данных
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}
	defer db.Close()

	// Создаем/обновляем таблицы
	if err := db.Initialize(); err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v", err)
	}

	log.Println("База данных успешно инициализирована")

	log.Println("FileMarker запущен")
	log.Printf("Используются следующие директории:")
	log.Printf("  Входящие: %s", cfg.IncomingDir)
	log.Printf("  Обработка: %s", cfg.ProcessingDir)
	log.Printf("  Исходящие: %s", cfg.OutgoingDir)
	log.Printf("  Архив: %s", cfg.ArchiveDir)

	// Создаем и запускаем веб-сервер
	server, err := web.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("Ошибка создания сервера: %v", err)
	}

	log.Printf("Запуск веб-сервера на порту %s", cfg.ServerPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// setupLogging настраивает логгирование в файл
func setupLogging(logsDir string) (*os.File, error) {
	logPath := filepath.Join(logsDir, "filemarker.log")

	// Создаем или открываем файл для логов
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Настраиваем вывод логов в файл и консоль
	log.SetOutput(logFile)

	return logFile, nil
}
