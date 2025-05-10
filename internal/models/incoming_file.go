package models

import (
	"time"
)

// IncomingFile представляет модель входящего файла для обработки
type IncomingFile struct {
	FileName    string    // Имя файла
	FilePath    string    // Полный путь к файлу
	GTIN        string    // Код товара (GTIN)
	Date        string    // Дата производства
	BatchNumber string    // Номер партии
	TotalCodes  int       // Общее количество кодов
	FileSize    int64     // Размер файла в байтах
	ModTime     time.Time // Время последнего изменения
}

// NewIncomingFile создает новый экземпляр входящего файла
func NewIncomingFile(fileName, filePath, gtin, date, batchNumber string, totalCodes int, fileSize int64, modTime time.Time) *IncomingFile {
	return &IncomingFile{
		FileName:    fileName,
		FilePath:    filePath,
		GTIN:        gtin,
		Date:        date,
		BatchNumber: batchNumber,
		TotalCodes:  totalCodes,
		FileSize:    fileSize,
		ModTime:     modTime,
	}
}
