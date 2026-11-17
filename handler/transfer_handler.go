package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"manisora/model"
	"manisora/service"
)

func transferStatusCode(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrTransferNotDraft),
		errors.Is(err, service.ErrTransferNotSent),
		errors.Is(err, service.ErrInsufficientStock):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func CreateTransfer(c *gin.Context) {
	var input model.Transfer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	t, err := service.CreateTransfer(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func GetTransfer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	t, err := service.GetTransferByID(id)
	if err != nil {
		c.JSON(transferStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func SendTransfer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	t, err := service.SendTransfer(id)
	if err != nil {
		c.JSON(transferStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func ReceiveTransfer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	var input model.ReceiveInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	t, err := service.ReceiveTransfer(id, input)
	if err != nil {
		c.JSON(transferStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}
