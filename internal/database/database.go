// internal/database/database.go
package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // Драйвер SQLite
)

// DB представляет соединение с базой данных
type DB struct {
	conn *sql.DB
}

// New создает новое подключение к базе данных
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Тестируем соединение
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

// Initialize создает необходимые таблицы, если они не существуют
func (db *DB) Initialize() error {
	// Таблица tasks
	_, err := db.conn.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		gtin TEXT NOT NULL,
		date TEXT NOT NULL,
		batch_number TEXT NOT NULL,
		total_codes INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		status TEXT NOT NULL,
		processed_at DATETIME,
		completed_at DATETIME,
		valid_codes INTEGER DEFAULT 0,
		invalid_codes INTEGER DEFAULT 0,
		duplicate_codes INTEGER DEFAULT 0
	)`)
	if err != nil {
		return err
	}

	// Таблица codes
	_, err = db.conn.Exec(`
	CREATE TABLE IF NOT EXISTS codes (
		value TEXT NOT NULL,
		task_id TEXT NOT NULL,
		scanned BOOLEAN DEFAULT 0,
		scanned_at DATETIME,
		valid BOOLEAN DEFAULT 1,
		error_message TEXT,
		PRIMARY KEY (value, task_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	)`)
	if err != nil {
		return err
	}

	// Таблица active_task
	_, err = db.conn.Exec(`
	CREATE TABLE IF NOT EXISTS active_task (
		task_id TEXT PRIMARY KEY,
		started_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	)`)
	if err != nil {
		return err
	}

	return nil
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	return db.conn.Close()
}
