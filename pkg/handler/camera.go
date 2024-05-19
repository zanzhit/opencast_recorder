package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	recorder "github.com/zanzhit/opencast_recorder"
)

func (h *Handler) create(c *gin.Context) {
	var input recorder.Camera
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := h.services.Create(input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	c.Status(http.StatusOK)
}
