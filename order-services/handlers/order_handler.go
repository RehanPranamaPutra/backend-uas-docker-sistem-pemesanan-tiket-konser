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

	// 1. AMBIL HARGA DARI LARAVEL (Catalog Service)
	laravelURL := fmt.Sprintf("http://catalog-service:8000/api/concerts/%d", order.EventID)
	respL, err := http.Get(laravelURL)
	if err != nil || respL.StatusCode != 200 {
		c.JSON(404, gin.H{"error": "Layanan Katalog tidak tersedia atau Konser tidak ditemukan"})
		return
	}
	var concert LaravelConcert
	json.NewDecoder(respL.Body).Decode(&concert)
	
	// Hitung Total (Keamanan: Backend yang menghitung harga, bukan frontend)
	order.Total = float64(order.Quantity) * concert.Price

	// 2. TANYA PYTHON (Reservation Service) - Kunci Stok di Redis
	// FINAL: Menggunakan %s di bagian akhir karena UserID adalah String MongoDB
	pythonURL := fmt.Sprintf("http://reservation-service:5002/reserve/%d/%d/%s", 
                  order.EventID, order.Quantity, order.UserID)
	respP, err := http.Post(pythonURL, "application/json", nil)
	
	if err != nil || respP.StatusCode != 200 {
		// Jika Python kirim 400/409, berarti stok habis atau user sedang ada lock aktif
		c.JSON(http.StatusConflict, gin.H{"error": "Stok tiket tidak mencukupi atau Anda memiliki pesanan yang belum dibayar"})
		return
	}

	// 3. SIMPAN STATUS PENDING KE POSTGRES
	order.Status = "PENDING"
	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database Postgres gagal menyimpan data"})
		return
	}

	c.JSON(201, gin.H{
		"message": "Pesanan berhasil dibuat (Status: PENDING)",
		"data":    order,
	})
}

func GetAllOrders(c *gin.Context) {
    var orders []models.Order
    
    // Mengambil semua data dari database (Select * from orders)
    // Sesuaikan 'database.DB' dengan variabel koneksi DB kamu
    if err := database.DB.Find(&orders).Error; err != nil {
        c.JSON(500, gin.H{"message": "Gagal mengambil data order", "error": err.Error()})
        return
    }

    // Return data sebagai JSON
    c.JSON(200, gin.H{"data": orders})
}

func ConfirmPayment(c *gin.Context) {
	id := c.Param("id")
	var order models.Order

	// 1. Cari data order di Postgres
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Data pesanan tidak ditemukan di database"})
		return
	}

	// 2. Update status di Postgres menjadi SUCCESS
	order.Status = "SUCCESS"
	database.DB.Save(&order)

	// 3. KONFIRMASI KE PYTHON (Hapus Lock di Redis karena sudah bayar)
	// FINAL: Pastikan mengirim 3 parameter (%d/%d/%s) agar sinkron dengan main.py
	confirmURL := fmt.Sprintf("http://reservation-service:5002/confirm-payment/%d/%d/%s", 
                  order.EventID, order.Quantity, order.UserID)
	
	// Kirim POST dengan body kosong
	http.Post(confirmURL, "application/json", strings.NewReader(""))

	// 4. UPDATE STOK PERMANEN DI LARAVEL (MySQL)
	laravelURL := fmt.Sprintf("http://catalog-service:8000/api/concerts/%d/stock", order.EventID)
	// Laravel menerima 'reduce_by' untuk mengurangi stok MySQL
	payload := strings.NewReader(fmt.Sprintf(`{"reduce_by": %d}`, order.Quantity))
	req, _ := http.NewRequest("PATCH", laravelURL, payload)
	req.Header.Add("Content-Type", "application/json")
	
	client := &http.Client{}
	client.Do(req)

	c.JSON(200, gin.H{"message": "Pembayaran Berhasil! Stok MySQL & Redis telah diperbarui."})
}

func GetUserOrders(c *gin.Context) {
	userId := c.Param("userId") // Ini akan menerima String MongoDB ID
	var orders []models.Order

	// Mencari berdasarkan UserID String
	if err := database.DB.Where("user_id = ?", userId).Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data tiket"})
		return
	}

	c.JSON(http.StatusOK, orders)
}