package models

import (
	"time"
	"gorm.io/gorm"
)

// Product represents a Product catalog for e-commerce
type Product struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name string `json:"name" gorm:"not null;uniqueIndex"` // Product name
	Description string `json:"description" gorm:"type:text"` // Product description
	Price float64 `json:"price" gorm:"type:decimal(10,2);not null"` // Product price
	SKU string `json:"sku" gorm:"not null;uniqueIndex"` // Stock Keeping Unit
	Stock int `json:"stock" gorm:"not null;default:0"` // Available stock quantity
	IsActive bool `json:"is_active" gorm:"default:true"` // Whether the product is active
	CategoryID uint `json:"category_id" gorm:"not null;index"` // Product category ID
}

// TableName returns the table name for Product
func (Product) TableName() string {
	return "products"
}
