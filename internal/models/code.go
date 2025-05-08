package models

import (
	"time"
)

// Code представляет код маркировки
type Code struct {
	Value        string    // Значение кода
	TaskID       string    // ID задания
	Scanned      bool      // Был ли код отсканирован
	ScannedAt    time.Time // Время сканирования
	Valid        bool      // Валидный ли код
	ErrorMessage string    // Сообщение об ошибке, если код невалидный
}

// NewCode создает новый экземпляр кода
func NewCode(value, taskID string) *Code {
	return &Code{
		Value:        value,
		TaskID:       taskID,
		Scanned:      false,
		Valid:        true, // По умолчанию считаем код валидным
		ErrorMessage: "",
	}
}
