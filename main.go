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
	r.POST("/goods-receipts", handler.CreateGoodsReceipt)
	r.POST("/items", handler.CreateItem)
	r.GET("/locations", handler.GetLocations)
	r.POST("/locations", handler.CreateLocation)
	r.GET("/suppliers", handler.GetSuppliers)
	r.POST("/suppliers", handler.CreateSupplier)
	r.GET("/goods-receipts", handler.GetGoodsReceipts)
	r.GET("/goods-receipts/:id", handler.GetGoodsReceipt)
	r.POST("/goods-receipts/:id/post", handler.PostGoodsReceipt)
	r.POST("/sales", handler.CreateSale)
	r.POST("/sales/:id/post", handler.PostSale)
	r.POST("/transfers", handler.CreateTransfer)
	r.GET("/transfers/:id", handler.GetTransfer)
	r.POST("/transfers/:id/send", handler.SendTransfer)
	r.POST("/transfers/:id/receive", handler.ReceiveTransfer)
	r.POST("/supplier-returns", handler.CreateSupplierReturn)
	r.GET("/supplier-returns/:id", handler.GetSupplierReturn)
	r.POST("/supplier-returns/:id/post", handler.PostSupplierReturn)
	r.POST("/stock-opnames", handler.CreateStockOpname)
	r.GET("/stock-opnames/:id", handler.GetStockOpname)
	r.PUT("/stock-opnames/:id/count", handler.SaveOpnameCount)
	r.POST("/stock-opnames/:id/approve", handler.ApproveStockOpname)
	r.Run(":8080")

}
