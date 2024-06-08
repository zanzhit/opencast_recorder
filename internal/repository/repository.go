package repository

import (
	"database/sql"

	"github.com/zanzhit/opencast_recorder/internal/domain/models"
)

type Camera interface {
	Create(camera models.Camera) error
}

type Recording interface {
	Start(rec models.Recording) error
	Stop(cameraIP string) error
	Move(cameraIP string) error
	LastRecording(cameraIP string) (models.Recording, error)
}

type Repository struct {
	Camera
	Recording
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Camera:    NewCameraSQLite(db),
		Recording: NewRecordingSQLite(db),
	}
}
