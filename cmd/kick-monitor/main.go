package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/retconned/kick-monitor/internal/api"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db.Init()

	var activeChannels []models.MonitoredChannel
	if err := db.DB.Where("is_active = ?", true).Find(&activeChannels).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("Failed to load active channels: %v", err)
		}
		log.Println("No active channels found in the database on startup.")
	}

	for _, channel := range activeChannels {
		go monitor.StartMonitoringChannel(&channel)
	}

	// Setup HTTP server
	http.HandleFunc("/add_channel", api.AddChannelHandler)
	http.HandleFunc("/process_livestream_report", api.ProcessLivestreamReportHandler)

	// New Routing for Report Endpoints
	http.HandleFunc("/reports/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/reports/") {
			api.GetReportByUUIDHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/channels/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports") {
			api.GetReportsByChannelIDHandler(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/info") {
			api.GetChannelInfoHandler(w, r)
		} else if r.URL.Path == "/channels" || r.URL.Path == "/channels/" { // Exactly /channels or /channels/
			api.GetMonitoredChannelsHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/livestreams/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reports") {
			api.GetReportsByLivestreamIDHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
