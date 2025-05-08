package main

import (
	"FileMarker/internal/filemanager"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: testparse <путь_к_xml_файлу>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Тестируем парсинг
	task, codes, err := filemanager.ParseMarkFile(filePath)
	if err != nil {
		log.Fatalf("Ошибка при парсинге файла: %v", err)
	}

	// Выводим информацию о задании
	fmt.Println("Информация о задании:")
	fmt.Printf("  ID: %s\n", task.ID)
	fmt.Printf("  GTIN: %s\n", task.GTIN)
	fmt.Printf("  Дата: %s\n", task.Date)
	fmt.Printf("  Номер партии: %s\n", task.BatchNumber)
	fmt.Printf("  Всего кодов: %d\n", task.TotalCodes)

	// Выводим первые 5 кодов
	fmt.Println("\nПримеры кодов:")
	for i, code := range codes {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s\n", i+1, code.Value)
	}
}
