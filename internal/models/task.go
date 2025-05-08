package models

import (
	"time"
)

// Константы статусов задания
const (
	TaskStatusNew        = "new"        // Новое задание
	TaskStatusProcessing = "processing" // В процессе обработки
	TaskStatusCompleted  = "completed"  // Завершено
)

// Task представляет задание по обработке файла XML
type Task struct {
	ID             string    // UUID документа из XML
	GTIN           string    // Код товара (GTIN)
	Date           string    // Дата производства
	BatchNumber    string    // Номер партии
	TotalCodes     int       // Общее количество кодов
	FilePath       string    // Путь к файлу OUT_MARK
	Status         string    // Статус задания
	ProcessedAt    time.Time // Время начала обработки
	CompletedAt    time.Time // Время завершения
	ValidCodes     int       // Количество валидных кодов
	InvalidCodes   int       // Количество невалидных кодов
	DuplicateCodes int       // Количество дубликатов
}

// NewTask создает новый экземпляр задания
func NewTask(id, gtin, date, batchNumber string, totalCodes int, filePath string) *Task {
	return &Task{
		ID:             id,
		GTIN:           gtin,
		Date:           date,
		BatchNumber:    batchNumber,
		TotalCodes:     totalCodes,
		FilePath:       filePath,
		Status:         TaskStatusNew,
		ValidCodes:     0,
		InvalidCodes:   0,
		DuplicateCodes: 0,
	}
}
