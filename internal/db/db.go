package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/retconned/kick-monitor/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
	)

	var err error
	for i := 0; i < 5; i++ { // Try up to 5 times
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break // Connection successful
		}
		log.Printf("Attempt %d: Failed to connect to database: %v. Retrying in 5 seconds...", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Exhausted retries: Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.MonitoredChannel{}, &models.ChannelData{}, &models.LivestreamData{}, &models.ChatMessage{}, &models.LivestreamReport{}, &models.SpamReport{}, &models.StreamerProfile{}, &models.User{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database schema: %v", err)
	}

	log.Println("Database connected and schema migrated.")
}
