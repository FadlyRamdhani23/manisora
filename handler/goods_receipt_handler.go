package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"manisora/model"
	"manisora/service"
)

func CreateGoodsReceipt(c *gin.Context) {
	var input model.GoodsReceipt
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format JSON tidak valid"})
		return
	}

	gr, err := service.CreateGoodsReceipt(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gr)
}
