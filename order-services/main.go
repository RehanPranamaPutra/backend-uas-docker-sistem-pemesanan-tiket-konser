package main

import (
	"order-services/database"
	"order-services/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Koneksi ke Database
	database.ConnectDB()

	// 2. Setup Router
	r := gin.Default()

	// 3. Endpoint Pesanan
	r.POST("/orders", handlers.CreateOrder)

	// 4. Jalankan Server di Port 8080
	r.Run(":8080")
}