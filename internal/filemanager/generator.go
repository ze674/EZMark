// internal/filemanager/generator.go
package filemanager

import (
	"FileMarker/internal/models"
	"encoding/xml"
	"os"
	"path/filepath"
)

// SerializationDocument структура для файла OUT_SERIALIZATION
type SerializationDocument struct {
	XMLName              xml.Name `xml:"Document"`
	DocumentID           string   `xml:"document_id"`
	GTIN                 string   `xml:"gitin"`
	Date                 string   `xml:"data"`
	Batch                string   `xml:"batch"`
	SerializationContent struct {
		CIS []struct {
			Value string `xml:",cdata"`
		} `xml:"cis"`
	} `xml:"serialization_content"`
}

// GenerateSerializationFile создает файл с результатами сериализации
func GenerateSerializationFile(task *models.Task, codes []*models.Code, outDir string) (string, error) {
	// Создаем структуру документа
	doc := SerializationDocument{}
	doc.DocumentID = task.ID
	doc.GTIN = task.GTIN
	doc.Date = task.Date
	doc.Batch = task.BatchNumber

	// Добавляем только отсканированные и валидные коды
	for _, code := range codes {
		if code.Scanned && code.Valid {
			cis := struct {
				Value string `xml:",cdata"`
			}{Value: code.Value}
			doc.SerializationContent.CIS = append(doc.SerializationContent.CIS, cis)
		}
	}

	// Создаем XML-документ
	xmlData, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	// Добавляем XML-заголовок
	xmlData = append([]byte(xml.Header), xmlData...)

	// Формируем имя файла
	fileName := "OUT_SERIALIZATION_" + task.ID + ".xml"
	filePath := filepath.Join(outDir, fileName)

	// Записываем файл
	err = os.WriteFile(filePath, xmlData, 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
