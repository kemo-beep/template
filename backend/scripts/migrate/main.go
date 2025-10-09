package main

import (
	"log"
	"mobile-backend/config"
	"mobile-backend/models"
)

func migrate() {
	// Connect to database
	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db := config.GetDB()

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Product{},
		&models.Plan{},
		&models.Subscription{},
		&models.Payment{},
		&models.PaymentMethod{},
		&models.WebhookEvent{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed successfully!")
}

func main() {
	migrate()
}
