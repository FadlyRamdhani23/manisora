package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"manisora/model"
	"manisora/service"
)

func opnameStatusCode(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrOpnameNotDraft),
		errors.Is(err, service.ErrInsufficientStock):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func CreateStockOpname(c *gin.Context) {
	var input model.StockOpname
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	o, err := service.CreateStockOpname(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func GetStockOpname(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	o, err := service.GetStockOpnameByID(id)
	if err != nil {
		c.JSON(opnameStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func SaveOpnameCount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	var input model.CountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	o, err := service.SaveOpnameCount(id, input)
	if err != nil {
		c.JSON(opnameStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func ApproveStockOpname(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	o, err := service.ApproveStockOpname(id)
	if err != nil {
		c.JSON(opnameStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}
