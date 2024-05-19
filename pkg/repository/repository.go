package repository

import (
	"database/sql"

	recorder "github.com/zanzhit/opencast_recorder"
)

type Camera interface {
	Create(camera recorder.Camera) error
}

type Recording interface {
	Start(rec recorder.Recording) error
	Stop(cameraIP string) error
	Move(cameraIP string) error
	LastRecording(cameraIP string) (recorder.Recording, error)
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
