package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"manisora/model"
	"manisora/service"
)

func GetLocations(c *gin.Context) {
	locations, err := service.GetAllLocations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func CreateLocation(c *gin.Context) {
	var input model.Location
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	loc, err := service.CreateLocation(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loc)
}
