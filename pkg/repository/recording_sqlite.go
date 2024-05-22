package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/errs"
)

type RecordingSQLite struct {
	db *sql.DB
}

func NewRecordingSQLite(db *sql.DB) *RecordingSQLite {
	return &RecordingSQLite{db: db}
}

func (r *RecordingSQLite) Start(rec recorder.Recording) error {
	rec.IsMoved = false

	query := fmt.Sprintf(`INSERT INTO %s (camera_ip, start_time, file_path, is_moved) VALUES (?, ?, ?, ?)`, recordingsTable)
	_, err := r.db.Exec(query, rec.CameraIP, rec.StartTime, rec.FilePath, rec.IsMoved)
	if err != nil {
		return fmt.Errorf("can't create new recording: %w", err)
	}

	return nil
}

func (r *RecordingSQLite) Stop(cameraIP string) error {
	stopTime := time.Now()
	fmt.Println(cameraIP)

	query := fmt.Sprintf(`UPDATE %s SET stop_time = ? WHERE record_id = (SELECT record_id FROM %s
						  WHERE camera_ip = ? ORDER BY start_time DESC LIMIT 1)`, recordingsTable, recordingsTable)
	_, err := r.db.Exec(query, stopTime, cameraIP)
	if err != nil {
		return fmt.Errorf("can't set stop time on recording: %w", err)
	}

	return nil
}

func (r *RecordingSQLite) Move(cameraIP string) error {
	isMoved := true
	query := fmt.Sprintf(`UPDATE %s SET is_moved = ? WHERE record_id = (SELECT record_id FROM %s
						  WHERE camera_ip = ? ORDER BY start_time DESC LIMIT 1)`, recordingsTable, recordingsTable)
	_, err := r.db.Exec(query, isMoved, cameraIP)
	if err != nil {
		return fmt.Errorf("can't change moved status: %w", err)
	}

	return nil
}

func (r *RecordingSQLite) LastRecording(cameraIP string) (recorder.Recording, error) {
	query := fmt.Sprintf(`SELECT camera_ip, start_time, stop_time, file_path, is_moved FROM %s WHERE record_id = (SELECT record_id FROM %s
						  WHERE camera_ip = ? ORDER BY start_time DESC LIMIT 1)`, recordingsTable, recordingsTable)

	var rec recorder.Recording
	err := r.db.QueryRow(query, cameraIP).Scan(&rec.CameraIP, &rec.StartTime, &rec.StopTime, &rec.FilePath, &rec.IsMoved)
	if err != nil {
		if err == sql.ErrNoRows {
			return rec, &errs.ErrNoRecording{}
		}
		return rec, fmt.Errorf("error retrieving recording: %w", err)
	}

	return rec, nil
}
