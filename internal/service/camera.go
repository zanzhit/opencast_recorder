package service

import (
	"github.com/zanzhit/opencast_recorder/internal/domain/models"
	"github.com/zanzhit/opencast_recorder/internal/repository"
)

type CameraService struct {
	repo repository.Camera
}

func NewCameraService(repo repository.Camera) *CameraService {
	return &CameraService{repo: repo}
}

func (r *CameraService) Create(camera models.Camera) error {
	if err := r.repo.Create(camera); err != nil {
		return err
	}

	return nil
}
