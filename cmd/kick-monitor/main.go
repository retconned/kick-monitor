package main

import (
	"log"
	"os"

	"github.com/retconned/kick-monitor/internal/api"
	"github.com/retconned/kick-monitor/internal/auth"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func main() {
	db.Init()

	// Initialize JWT Secret for authentication
	auth.InitAuth()

	// Start monitoring Go routines for active channels
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

	e := echo.New()

	// --- Global Middleware ---
	e.Use(middleware.Logger())
	e.Use(middleware.Recover()) // Recovers from panics and serves a 500 error

	// public routes start here
	e.POST("/register", auth.RegisterHandler)
	e.POST("/login", auth.LoginHandler)

	e.POST("/process_livestream_report", api.ProcessLivestreamReportHandler) // This is asynchronous, can be public

	// Reports API
	// e.GET("/reports/:reportUUID", api.GetReportByUUIDHandler) - specific route for single UUID
	// Group these routes with common prefixes
	e.GET("/reports/:reportUUID", api.GetReportByUUIDHandler)                        // /reports/uuid-string
	e.GET("/channels/:channelID/reports", api.GetReportsByChannelIDHandler)          // /channels/id/reports
	e.GET("/livestreams/:livestreamID/reports", api.GetReportsByLivestreamIDHandler) // /livestreams/id/reports

	// Channels Info API
	// e.GET("/channels", api.GetMonitoredChannelsHandler) // /channels (list all monitored)
	// e.GET("/channels/:channelID/info", api.GetChannelInfoHandler) // /channels/id/info (latest raw data snapshot) // i don't think this should be a thing
	// e.GET("/channels/:channelID/profile", api.GetStreamerProfileHandler) // /channels/id/profile (aggregated profile)

	// proeteced routes start here
	r := e.Group("/protected")
	r.Use(auth.AuthMiddleware())
	r.POST("/add_channel", api.AddChannelHandler) // /protected/add_channel

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("Starting server on :%s", port)
	e.Logger.Fatal(e.Start(":" + port)) // Use e.Logger.Fatal for Echo's logger
}
