package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"manisora/model"
	"manisora/service"
)

func returnStatusCode(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrNotDraft),
		errors.Is(err, service.ErrReturnExceedsReceipt),
		errors.Is(err, service.ErrInsufficientStock):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func CreateSupplierReturn(c *gin.Context) {
	var input model.SupplierReturn
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	r, err := service.CreateSupplierReturn(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func GetSupplierReturn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	r, err := service.GetSupplierReturnByID(id)
	if err != nil {
		c.JSON(returnStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func PostSupplierReturn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	r, err := service.PostSupplierReturn(id)
	if err != nil {
		c.JSON(returnStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func GetSupplierReturns(c *gin.Context) {
	list, err := service.GetAllSupplierReturns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
