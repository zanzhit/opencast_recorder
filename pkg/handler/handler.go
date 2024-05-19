package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zanzhit/opencast_recorder/pkg/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	cameras := router.Group("/cameras")
	{
		cameras.POST("/:camera_ip/start", h.start)
		cameras.POST("/:camera_ip/stop", h.stop)
		cameras.POST("/:camera_ip/move", h.move)
		cameras.POST("/:camera_ip/schedule", h.schedule)
		cameras.POST("", h.create)

		cameras.GET("/:camera_ip", h.stats)
	}

	return router
}
