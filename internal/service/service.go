package service

import (
	"net/http"

	"github.com/zanzhit/opencast_recorder/internal/domain/models"
	"github.com/zanzhit/opencast_recorder/internal/repository"
)

type Camera interface {
	Create(camera models.Camera) error
}

type Recording interface {
	Start(rec []models.Recording) error
	Stop(cameraIP string) error
	Move(string) (*http.Response, error)
	DeleteLocal(string) error
	Schedule(rec models.RecordingSchedule) error
	Stats(string) (models.Recording, error)
}

type Service struct {
	Camera
	Recording
}

func NewService(repos *repository.Repository, video VideoService, videoPath string) *Service {
	return &Service{
		Camera:    NewCameraService(repos.Camera),
		Recording: NewRecordingService(repos.Recording, video, videoPath),
	}
}
