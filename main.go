package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"manisora/database"
	"manisora/handler"
)

func main() {
	database.ConnectDB()
	defer database.DB.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))

	// Master data
	r.GET("/items", handler.GetItems)
	r.POST("/items", handler.CreateItem)
	r.GET("/locations", handler.GetLocations)
	r.POST("/locations", handler.CreateLocation)
	r.GET("/suppliers", handler.GetSuppliers)
	r.POST("/suppliers", handler.CreateSupplier)

	// Penerimaan dari supplier
	r.POST("/goods-receipts", handler.CreateGoodsReceipt)
	r.GET("/goods-receipts", handler.GetGoodsReceipts)
	r.GET("/goods-receipts/:id", handler.GetGoodsReceipt)
	r.POST("/goods-receipts/:id/post", handler.PostGoodsReceipt)

	// Penjualan
	r.POST("/sales", handler.CreateSale)
	r.GET("/sales", handler.GetSales)
	r.GET("/sales/:id", handler.GetSale)
	r.POST("/sales/:id/post", handler.PostSale)

	// Transfer ke unit/gudang
	r.POST("/transfers", handler.CreateTransfer)
	r.GET("/transfers", handler.GetTransfers)
	r.GET("/transfers/:id", handler.GetTransfer)
	r.POST("/transfers/:id/send", handler.SendTransfer)
	r.POST("/transfers/:id/receive", handler.ReceiveTransfer)

	// Retur ke supplier
	r.POST("/supplier-returns", handler.CreateSupplierReturn)
	r.GET("/supplier-returns", handler.GetSupplierReturns)
	r.GET("/supplier-returns/:id", handler.GetSupplierReturn)
	r.POST("/supplier-returns/:id/post", handler.PostSupplierReturn)

	// Opname
	r.POST("/stock-opnames", handler.CreateStockOpname)
	r.GET("/stock-opnames", handler.GetStockOpnames)
	r.GET("/stock-opnames/:id", handler.GetStockOpname)
	r.PUT("/stock-opnames/:id/count", handler.SaveOpnameCount)
	r.POST("/stock-opnames/:id/approve", handler.ApproveStockOpname)

	// Laporan
	r.GET("/reports/stock-mutation", handler.GetStockMutationReport)

	log.Println("Starting server on :8080")
	r.Run(":8080")
}
