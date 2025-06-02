package main

import (
	"log"
	"net/http"

	"github.com/retconned/kick-monitor/internal/api"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"
)

func main() {
	// Initialize database
	db.Init()

	// Load active channels and start monitoring
	var activeChannels []models.MonitoredChannel
	if err := db.DB.Where("is_active = ?", true).Find(&activeChannels).Error; err != nil {
		log.Fatalf("Failed to load active channels: %v", err)
	}

	for i := range activeChannels {
		// Use a separate variable for the goroutine to avoid capturing the loop variable
		channel := activeChannels[i]
		go monitor.StartMonitoringChannel(&channel)
	}

	// Setup HTTP server
	http.HandleFunc("/add_channel", api.AddChannelHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
