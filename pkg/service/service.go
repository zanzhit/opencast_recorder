package service

import (
	"net/http"

	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/repository"
)

type Camera interface {
	Create(camera recorder.Camera) error
}

type Recording interface {
	Start(rec []recorder.Recording) error
	Stop(cameraIP string) error
	Move(string) (*http.Response, error)
	Schedule(rec recorder.RecordingSchedule) error
	Stats(string) (recorder.Recording, error)
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
