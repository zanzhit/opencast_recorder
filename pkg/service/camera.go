package service

import (
	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/repository"
)

type CameraService struct {
	repo repository.Camera
}

func NewCameraService(repo repository.Camera) *CameraService {
	return &CameraService{repo: repo}
}

func (r *CameraService) Create(camera recorder.Camera) error {
	if err := r.repo.Create(camera); err != nil {
		return err
	}

	return nil
}
