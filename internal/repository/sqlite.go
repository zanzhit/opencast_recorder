package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	camerasTable    = "cameras"
	recordingsTable = "recordings"
)

func NewSQLiteDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("can't create tables %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		camera_ip TEXT PRIMARY KEY,
		room_number TEXT NOT NULL,
		has_audio BOOLEAN NOT NULL
	);

	CREATE TABLE IF NOT EXISTS %s (
		record_id INTEGER PRIMARY KEY,
		camera_ip TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		stop_time DATETIME,
		file_path TEXT NOT NULL,
		is_moved BOOLEAN NOT NULL,
		FOREIGN KEY (camera_ip) REFERENCES cameras (camera_ip)
	);
	`, camerasTable, recordingsTable)
	_, err := db.Exec(query)
	return err
}
