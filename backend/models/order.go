package models

import (
	"time"
	"gorm.io/gorm"
)

// Order represents a Customer orders for e-commerce
type Order struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	OrderNumber string `json:"order_number" gorm:"not null;uniqueIndex"` // Unique order number
	CustomerID uint `json:"customer_id" gorm:"not null;index"` // Customer ID
	TotalAmount float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"` // Total order amount
	Status string `json:"status" gorm:"not null;default:'pending'"` // Order status
	Notes string `json:"notes" gorm:"type:text"` // Order notes
}

// TableName returns the table name for Order
func (Order) TableName() string {
	return "orders"
}
