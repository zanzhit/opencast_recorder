package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zanzhit/opencast_recorder/internal/domain/models"
	"github.com/zanzhit/opencast_recorder/internal/errs"
)

func (h *Handler) start(c *gin.Context) {
	ips := strings.Split(c.Param("camera_ip"), ",")

	cameras := make([]models.Recording, len(ips))
	for i := 0; i < len(ips); i++ {
		cameras[i].CameraIP = ips[i]
		cameras[i].RTSP = fmt.Sprintf("rtsp://%s", ips[i])
	}

	if err := h.services.Start(cameras); err != nil {
		switch err.(type) {
		case *errs.BadRequst:
			newErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.Status(http.StatusCreated)
}

// Only the last ip is taken for recording, several values separated by commas are taken in transit only for convenience of the user.
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

	response, err := h.services.Move(selectedRecord)
	if err != nil {
		switch err.(type) {
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	c.JSON(response.StatusCode, responseBody)
}

func (h *Handler) move(c *gin.Context) {
	ips := strings.Split(c.Param("camera_ip"), ",")
	selectedRecord := ips[len(ips)-1]

	response, err := h.services.Move(selectedRecord)
	if err != nil {
		switch err.(type) {
		case *errs.ErrNoRecording:
			newErrorResponse(c, http.StatusNotFound, err.Error())
			return
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		rec, err := h.services.Stats(selectedRecord)
		if err != nil {
			switch err.(type) {
			case *errs.ErrNoRecording:
				newErrorResponse(c, http.StatusNotFound, err.Error())
			default:
				newErrorResponse(c, http.StatusInternalServerError, err.Error())
			}
		}
		if err := h.services.DeleteLocal(rec.FilePath); err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(response.StatusCode, responseBody)
}

func (h *Handler) schedule(c *gin.Context) {
	var input models.RecordingSchedule
	ips := strings.Split(c.Param("camera_ip"), ",")
	for i := 0; i < len(ips); i++ {
		input.Recordings = append(input.Recordings, models.Recording{CameraIP: ips[i], RTSP: fmt.Sprintf("rtsp://%s", ips[i])})
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := scheduleDurationToInt(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid duration")
		return
	}

	go func() {
		if err := h.services.Schedule(input); err != nil {
			logrus.Print("schedule error: ", err.Error())
		}

		selectedRecord := ips[len(ips)-1]
		_, err := h.services.Move(selectedRecord)
		if err != nil {
			logrus.Print("moving error: ", err.Error())
		}
	}()

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

func scheduleDurationToInt(rec *models.RecordingSchedule) error {
	parts := strings.Split(rec.DurationStr, ":")
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}

	rec.Duration = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

	return nil
}
