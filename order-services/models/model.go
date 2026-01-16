package models

import "time"

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`    // UBAH DARI uint KE string
	EventID   uint      `json:"event_id"`   
	Quantity  int       `json:"quantity"`   
	Total     float64   `json:"total"`      
	Status    string    `gorm:"default:'PENDING'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}