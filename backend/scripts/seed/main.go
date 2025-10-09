package main

import (
	"fmt"
	"log"
	"mobile-backend/config"
	"mobile-backend/models"

	"golang.org/x/crypto/bcrypt"
)

func seed() {
	// Connect to database
	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db := config.GetDB()

	// Create admin user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.User{
		Email:    "admin@example.com",
		Password: string(hashedPassword),
		Name:     "Admin User",
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Admin user already exists or error: %v", err)
	} else {
		log.Println("Admin user created successfully")
	}

	// Create test users
	for i := 1; i <= 10; i++ {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := models.User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: string(hashedPassword),
			Name:     fmt.Sprintf("Test User %d", i),
			IsActive: true,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("User %d already exists or error: %v", i, err)
		} else {
			log.Printf("User %d created successfully", i)
		}
	}

	log.Println("Database seeded successfully!")
}

func main() {
	seed()
}
