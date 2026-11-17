package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"manisora/service"
)

func GetStockMutationReport(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Query("location_id"), 10, 64)
	if err != nil || locationID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location_id wajib diisi"})
		return
	}

	var itemID int64
	if v := c.Query("item_id"); v != "" {
		itemID, err = strconv.ParseInt(v, 10, 64)
		if err != nil || itemID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_id tidak valid"})
			return
		}
	}

	report, err := service.GetStockMutationReport(locationID, c.Query("from"), c.Query("to"), itemID)
	if errors.Is(err, service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
