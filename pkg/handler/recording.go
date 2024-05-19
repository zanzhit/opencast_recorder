package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/errs"
)

func (h *Handler) start(c *gin.Context) {
	ips := strings.Split(c.Param("camera_ip"), ",")

	cameras := make([]recorder.Recording, len(ips))
	for i := 0; i < len(ips); i++ {
		cameras[i].CameraIP = ips[i]
		cameras[i].RTSP = fmt.Sprintf("rtsp://%s", ips[i])
	}

	if err := h.services.Start(cameras); err != nil {
		switch err.(type) {
		case *errs.BadRequst:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
	}

	c.Status(http.StatusCreated)
}

// Only the last ipi is taken for recording, several values separated by commas are taken in transit only for convenience of the user.
func (h *Handler) stop(c *gin.Context) {
	ips := strings.Split(c.Param("camera_ip"), ",")
	selectedRecord := ips[len(ips)-1]

	if err := h.services.Stop(selectedRecord); err != nil {
		switch err.(type) {
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
	}

	c.Status(http.StatusOK)
}

func (h *Handler) move(c *gin.Context) {
	ips := strings.Split(c.Param("camera_ip"), ",")
	selectedRecord := ips[len(ips)-1]

	respBody, err := h.services.Move(selectedRecord)
	if err != nil {
		switch err.(type) {
		case *errs.BadRequst:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		default:
			c.JSON(http.StatusInternalServerError, respBody)
			// newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(http.StatusCreated, respBody)
}

func (h *Handler) schedule(c *gin.Context) {
	var input recorder.RecordingSchedule
	ips := strings.Split(c.Param("camera_ip"), ",")
	for i := 0; i < len(ips); i++ {
		input.Recordings = append(input.Recordings, recorder.Recording{CameraIP: ips[i], RTSP: fmt.Sprintf("rtsp://%s", ips[i])})
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := h.services.Schedule(input); err != nil {
		switch err.(type) {
		case *errs.BadRequst:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *Handler) stats(c *gin.Context) {
	recording, err := h.services.Stats(c.Param("camera_ip"))
	if err != nil {
		switch err.(type) {
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
	}

	c.JSON(http.StatusOK, recording)
}
