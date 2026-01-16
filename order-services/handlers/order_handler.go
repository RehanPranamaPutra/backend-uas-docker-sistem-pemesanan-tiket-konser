package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order-services/database"
	"order-services/models"
	"strings"
	"github.com/gin-gonic/gin"
)

// Struktur bantuan untuk membaca respon dari Laravel
type LaravelConcert struct {
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	// 1. AMBIL HARGA DARI LARAVEL
	laravelURL := fmt.Sprintf("http://catalog-service:8000/api/concerts/%d", order.EventID)
	respL, err := http.Get(laravelURL)
	if err != nil || respL.StatusCode != 200 {
		c.JSON(404, gin.H{"error": "Konser tidak ditemukan di Laravel"})
		return
	}
	var concert LaravelConcert
	json.NewDecoder(respL.Body).Decode(&concert)
	order.Total = float64(order.Quantity) * concert.Price

	// 2. TANYA PYTHON (RESERVE STOK)
	// Pastikan URL menyertakan user_id sesuai main.py Python
	pythonURL := fmt.Sprintf("http://reservation-service:5002/reserve/%d/%d/%d", 
                  order.EventID, order.Quantity, order.UserID)
	respP, err := http.Post(pythonURL, "application/json", nil)
	
	if err != nil || respP.StatusCode != 200 {
		c.JSON(409, gin.H{"error": "Stok habis atau sedang dikunci"})
		return
	}

	// 3. SIMPAN PENDING KE POSTGRES
	order.Status = "PENDING"
	database.DB.Create(&order)

	c.JSON(201, gin.H{"message": "Pesanan dibuat (PENDING)", "data": order})
}

func ConfirmPayment(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Order not found"})
		return
	}

	// 1. UPDATE POSTGRES
	order.Status = "SUCCESS"
	database.DB.Save(&order)

	// 2. KONFIRMASI KE PYTHON (Hapus Lock)
	confirmURL := fmt.Sprintf("http://reservation-service:5002/confirm-payment/%d/%d", 
                   order.EventID, order.UserID)
	http.Post(confirmURL, "application/json", nil)

	// 3. UPDATE STOK PERMANEN DI LARAVEL (MySQL)
	patchURL := fmt.Sprintf("http://catalog-service:8000/api/concerts/%d/stock", order.EventID)
	payload := strings.NewReader(fmt.Sprintf(`{"reduce_by": %d}`, order.Quantity))
	req, _ := http.NewRequest("PATCH", patchURL, payload)
	req.Header.Add("Content-Type", "application/json")
	
	client := &http.Client{}
	client.Do(req)

	c.JSON(200, gin.H{"message": "Pembayaran Berhasil! Stok Sinkron."})
}