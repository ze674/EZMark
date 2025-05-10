// internal/database/codes.go
package database

import (
	"FileMarker/internal/models"
	"database/sql"
	"time"
)

// SaveCodes сохраняет коды в базу данных
func (db *DB) SaveCodes(codes []*models.Code) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
	INSERT INTO codes (value, task_id, scanned, valid, error_message)
	VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, code := range codes {
		_, err := stmt.Exec(code.Value, code.TaskID, code.Scanned, code.Valid, code.ErrorMessage)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetCodesByTaskID получает все коды для задания
func (db *DB) GetCodesByTaskID(taskID string) ([]*models.Code, error) {
	rows, err := db.conn.Query(`
	SELECT value, task_id, scanned, scanned_at, valid, error_message
	FROM codes WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*models.Code
	for rows.Next() {
		code := &models.Code{}
		var scannedAt sql.NullTime // Для обработки NULL значений

		err := rows.Scan(&code.Value, &code.TaskID, &code.Scanned, &scannedAt, &code.Valid, &code.ErrorMessage)
		if err != nil {
			return nil, err
		}

		if scannedAt.Valid {
			code.ScannedAt = scannedAt.Time
		}

		codes = append(codes, code)
	}

	return codes, nil
}

// UpdateCodeStatus обновляет статус кода (отсканирован, валидный и т.д.)
func (db *DB) UpdateCodeStatus(taskID, codeValue string, scanned, valid bool, errorMsg string) error {
	// Если код отсканирован, устанавливаем время сканирования
	var scannedAt interface{}
	if scanned {
		scannedAt = time.Now()
	} else {
		scannedAt = nil
	}

	_, err := db.conn.Exec(`
	UPDATE codes SET 
	scanned = ?, scanned_at = ?, valid = ?, error_message = ?
	WHERE task_id = ? AND value = ?`,
		scanned, scannedAt, valid, errorMsg, taskID, codeValue)

	return err
}

// GetScannedCodesByTaskID получает все отсканированные коды для задания
func (db *DB) GetScannedCodesByTaskID(taskID string) ([]*models.Code, error) {
	rows, err := db.conn.Query(`
	SELECT value, task_id, scanned, scanned_at, valid, error_message
	FROM codes WHERE task_id = ? AND scanned = 1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*models.Code
	for rows.Next() {
		code := &models.Code{}
		var scannedAt sql.NullTime // Для обработки NULL значений

		err := rows.Scan(&code.Value, &code.TaskID, &code.Scanned, &scannedAt, &code.Valid, &code.ErrorMessage)
		if err != nil {
			return nil, err
		}

		if scannedAt.Valid {
			code.ScannedAt = scannedAt.Time
		}

		codes = append(codes, code)
	}

	return codes, nil
}
