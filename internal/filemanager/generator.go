package filemanager

import (
	"FileMarker/internal/models"
	"encoding/xml"
	"os"
	"path/filepath"
)

// SerializationDocument представляет структуру XML-файла OUT_SERIALIZATION
type SerializationDocument struct {
	XMLName              xml.Name `xml:"Document"`
	DocumentID           string   `xml:"document_id"`
	GTIN                 string   `xml:"gitin"`
	Date                 string   `xml:"data"`
	BatchNumber          string   `xml:"batch"`
	SerializationContent struct {
		CIS []struct {
			Value string `xml:",cdata"`
		} `xml:"cis"`
	} `xml:"serialization_content"`
}

// GenerateSerializationFile создает файл OUT_SERIALIZATION_*.xml
func GenerateSerializationFile(task *models.Task, validCodes []string, outgoingDir string) (string, error) {
	// Создаем структуру документа
	doc := SerializationDocument{
		DocumentID:  task.ID,
		GTIN:        task.GTIN,
		Date:        task.Date,
		BatchNumber: task.BatchNumber,
	}

	// Добавляем валидные коды
	doc.SerializationContent.CIS = make([]struct {
		Value string `xml:",cdata"`
	}, len(validCodes))

	for i, code := range validCodes {
		doc.SerializationContent.CIS[i].Value = code
	}

	// Создаем имя файла
	fileName := "OUT_SERIALIZATION_" + task.ID + ".xml"
	filePath := filepath.Join(outgoingDir, fileName)

	// Создаем файл
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Записываем XML-заголовок
	file.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")

	// Создаем кодировщик XML
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	// Записываем документ
	if err := encoder.Encode(doc); err != nil {
		return "", err
	}

	return filePath, nil
}
