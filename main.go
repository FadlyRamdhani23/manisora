package main

import (
	"manisora/database"
	"manisora/handler"

	"github.com/gin-gonic/gin"
)

func main() {

	database.ConnectDB()
	r := gin.Default()
	r.GET("/items", handler.GetItems)
	r.POST("/items", handler.CreateItem)
	r.GET("/locations", handler.GetLocations)
	r.POST("/locations", handler.CreateLocation)
	r.GET("/suppliers", handler.GetSuppliers)
	r.POST("/suppliers", handler.CreateSupplier)
	r.Run(":8080")

}
