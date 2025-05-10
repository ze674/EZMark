// internal/database/tasks.go
package database

import (
	"FileMarker/internal/models"
	"database/sql"
	"time"
)

// SaveTask сохраняет задание в базу данных
func (db *DB) SaveTask(task *models.Task) error {
	_, err := db.conn.Exec(`
	INSERT INTO tasks (
		id, gtin, date, batch_number, total_codes, file_path, status
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.GTIN, task.Date, task.BatchNumber, task.TotalCodes, task.FilePath, task.Status)
	return err
}

// GetTaskByID получает задание по ID
func (db *DB) GetTaskByID(id string) (*models.Task, error) {
	task := &models.Task{ID: id}

	// Используем sql.NullTime для обработки NULL значений
	var processedAt sql.NullTime
	var completedAt sql.NullTime

	err := db.conn.QueryRow(`
    SELECT gtin, date, batch_number, total_codes, file_path, status, 
           processed_at, completed_at, valid_codes, invalid_codes, duplicate_codes 
    FROM tasks WHERE id = ?`, id).Scan(
		&task.GTIN, &task.Date, &task.BatchNumber, &task.TotalCodes,
		&task.FilePath, &task.Status, &processedAt, &completedAt,
		&task.ValidCodes, &task.InvalidCodes, &task.DuplicateCodes)

	if err != nil {
		return nil, err
	}

	// Конвертируем sql.NullTime в time.Time
	if processedAt.Valid {
		task.ProcessedAt = processedAt.Time
	}

	if completedAt.Valid {
		task.CompletedAt = completedAt.Time
	}

	return task, nil
}

// UpdateTaskStatus обновляет статус задания
func (db *DB) UpdateTaskStatus(id, status string) error {
	var query string
	var args []interface{}

	if status == models.TaskStatusProcessing {
		query = "UPDATE tasks SET status = ?, processed_at = ? WHERE id = ?"
		args = append(args, status, time.Now(), id)
	} else if status == models.TaskStatusCompleted {
		query = "UPDATE tasks SET status = ?, completed_at = ? WHERE id = ?"
		args = append(args, status, time.Now(), id)
	} else {
		query = "UPDATE tasks SET status = ? WHERE id = ?"
		args = append(args, status, id)
	}

	_, err := db.conn.Exec(query, args...)
	return err
}

// SetActiveTask устанавливает активное задание
func (db *DB) SetActiveTask(taskID string) error {
	// Сначала очищаем текущие активные задания
	_, err := db.conn.Exec("DELETE FROM active_task")
	if err != nil {
		return err
	}

	// Затем добавляем новое активное задание
	_, err = db.conn.Exec("INSERT INTO active_task (task_id, started_at) VALUES (?, ?)",
		taskID, time.Now())
	return err
}

// GetActiveTask получает текущее активное задание
func (db *DB) GetActiveTask() (string, error) {
	var taskID string
	err := db.conn.QueryRow("SELECT task_id FROM active_task LIMIT 1").Scan(&taskID)
	return taskID, err
}

// ClearActiveTask удаляет текущее активное задание
func (db *DB) ClearActiveTask() error {
	_, err := db.conn.Exec("DELETE FROM active_task")
	return err
}

// GetTaskStatistics обновляет статистику по заданию
func (db *DB) UpdateTaskStatistics(taskID string) error {
	// Подсчитываем статистику по кодам
	var validCodes, invalidCodes, duplicateCodes int

	// Получаем количество валидных отсканированных кодов
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM codes 
		WHERE task_id = ? AND scanned = 1 AND valid = 1`, taskID).Scan(&validCodes)
	if err != nil {
		return err
	}

	// Получаем количество невалидных кодов
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM codes 
		WHERE task_id = ? AND valid = 0`, taskID).Scan(&invalidCodes)
	if err != nil {
		return err
	}

	// Получаем количество дубликатов (для будущей реализации)
	// В текущей версии просто ставим 0
	duplicateCodes = 0

	// Обновляем статистику в задании
	_, err = db.conn.Exec(`
		UPDATE tasks SET 
		valid_codes = ?, invalid_codes = ?, duplicate_codes = ?
		WHERE id = ?`, validCodes, invalidCodes, duplicateCodes, taskID)

	return err
}
