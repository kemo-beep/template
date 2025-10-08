package models

import (
	"time"
	"gorm.io/gorm"
)

// Category represents a Product categories for organization
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name string `json:"name" gorm:"not null;uniqueIndex"` // Category name
	Description string `json:"description" gorm:"type:text"` // Category description
	Slug string `json:"slug" gorm:"not null;uniqueIndex"` // URL-friendly category identifier
	ParentID *uint `json:"parent_id,omitempty" gorm:"index"` // Parent category ID for hierarchical structure
	IsActive bool `json:"is_active" gorm:"default:true"` // Whether the category is active
	SortOrder int `json:"sort_order" gorm:"default:0"` // Sort order for display
}

// TableName returns the table name for Category
func (Category) TableName() string {
	return "categories"
}
