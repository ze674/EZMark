package filemanager

import (
	"FileMarker/internal/models"
	"encoding/xml"
	"os"
	"path/filepath"
)

// Обновленная структура MarkDocument для вашего упрощенного XML
type MarkDocument struct {
	XMLName      xml.Name `xml:"root"`
	DocumentID   string   `xml:"document_id"`
	GTIN         string   `xml:"gtin"`
	Date         string   `xml:"data"`
	BatchNumber  string   `xml:"batch"`
	CodeDivision struct {
		L00All int `xml:"l_00_all"` // Теперь это int без пробела
	} `xml:"code_division"`
	Labels struct {
		Label []string `xml:"label"`
	} `xml:"labels"`
}

// ParseMarkFile парсит XML-файл OUT_MARK и возвращает модель задания и список кодов
func ParseMarkFile(filePath string) (*models.Task, []*models.Code, error) {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	// Декодирование XML
	var doc MarkDocument
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&doc); err != nil {
		return nil, nil, err
	}

	// Получаем количество кодов
	totalCodes := len(doc.Labels.Label)

	// Создание модели задания
	task := models.NewTask(
		doc.DocumentID,
		doc.GTIN,
		doc.Date,
		doc.BatchNumber,
		totalCodes,
		filePath,
	)

	// Создание списка кодов
	var codes []*models.Code
	for _, label := range doc.Labels.Label {
		code := models.NewCode(label, doc.DocumentID)
		codes = append(codes, code)
	}

	return task, codes, nil
}

// MoveToProcessing перемещает файл из incoming в processing
func MoveToProcessing(filePath, processingDir string) (string, error) {
	fileName := filepath.Base(filePath)
	destPath := filepath.Join(processingDir, fileName)

	// Копируем сначала файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", err
	}

	// Удаляем исходный файл
	if err := os.Remove(filePath); err != nil {
		return "", err
	}

	return destPath, nil
}

// MoveToArchive перемещает файл из processing в archive
func MoveToArchive(filePath, archiveDir string) error {
	fileName := filepath.Base(filePath)
	destPath := filepath.Join(archiveDir, fileName)

	// Копируем сначала файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return err
	}

	// Удаляем исходный файл
	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}
