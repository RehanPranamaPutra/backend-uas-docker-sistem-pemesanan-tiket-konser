package handlers

import (
	"net/http"
	"order-services/database"
	"order-services/models"

	"github.com/gin-gonic/gin"
)

func CreateOrder(c *gin.Context) {
	var order models.Order

	// 1. Tangkap data JSON dari Flutter
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	// 2. Set status awal
	order.Status = "SUCCESS"

	// 3. Simpan ke PostgreSQL
	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pesanan"})
		return
	}

	// 4. Beri respon sukses
	c.JSON(http.StatusCreated, gin.H{
		"message": "Pesanan berhasil dibuat",
		"data":    order,
	})
}