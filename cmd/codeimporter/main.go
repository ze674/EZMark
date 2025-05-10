package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Запрос ID задания у пользователя
	var taskID int
	fmt.Print("Введите ID задания: ")
	fmt.Scan(&taskID)

	// Подключение к базе данных
	db, err := sql.Open("sqlite3", "./data/ezline.db")
	if err != nil {
		fmt.Printf("Ошибка подключения к базе данных: %v\n", err)
		return
	}
	defer db.Close()

	// Выполнение запроса
	rows, err := db.Query("SELECT code FROM items WHERE task_id = ?", taskID)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса: %v\n", err)
		return
	}
	defer rows.Close()

	// Сбор кодов
	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			fmt.Printf("Ошибка чтения результатов: %v\n", err)
			return
		}
		codes = append(codes, code)
	}

	// Проверка на наличие кодов
	if len(codes) == 0 {
		fmt.Printf("Для задания с ID %d не найдено кодов\n", taskID)
		return
	}

	// Сохранение в CSV
	filename := fmt.Sprintf("task_%d_codes.csv", taskID)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		return
	}
	defer file.Close()

	// Запись кодов построчно
	content := strings.Join(codes, "\n")
	if _, err := file.WriteString(content); err != nil {
		fmt.Printf("Ошибка записи в файл: %v\n", err)
		return
	}

	fmt.Printf("Успешно сохранено %d кодов в файл %s\n", len(codes), filename)
}
