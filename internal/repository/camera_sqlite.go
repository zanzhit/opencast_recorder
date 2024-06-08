package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zanzhit/opencast_recorder/internal/domain/models"
)

type CameraSQLite struct {
	db *sql.DB
}

func NewCameraSQLite(db *sql.DB) *CameraSQLite {
	return &CameraSQLite{db: db}
}

func (r *CameraSQLite) Create(camera models.Camera) error {
	query := fmt.Sprintf("INSERT INTO %s (camera_ip, room_number, has_audio) VALUES (?, ?, ?)", camerasTable)
	_, err := r.db.Exec(query, camera.CameraIP, camera.RoomNumber, camera.HasAudio)
	if err != nil {
		return fmt.Errorf("can't add new camera: %w", err)
	}

	return nil
}
