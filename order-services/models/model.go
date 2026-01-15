package models

import "time"

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`    // ID User dari Express
	EventID   uint      `json:"event_id"`   // ID Konser dari Laravel
	Quantity  int       `json:"quantity"`   // Jumlah tiket
	Total     float64   `json:"total"`      // Total harga
	Status    string    `json:"status"`     // PENDING / SUCCESS
	CreatedAt time.Time `json:"created_at"`
}