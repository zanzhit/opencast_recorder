package service

import (
	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Camera interface {
	Create(camera recorder.Camera) error
}

type Recording interface {
	Start(rec []recorder.Recording) error
	Stop(cameraIP string) error
	Move(string) ([]byte, error)
	Schedule(rec recorder.RecordingSchedule) error
	Stats(string) (recorder.Recording, error)
}

type Config struct {
	ACL          []byte
	Processing   []byte
	VideosPath   string
	VideoService string
	Login        string
	Password     string
}

type Service struct {
	Camera
	Recording
	Config
}

func NewService(repos *repository.Repository, cfg Config) *Service {
	return &Service{
		Camera:    NewCameraService(repos.Camera),
		Recording: NewRecordingService(repos.Recording, cfg),
	}
}
