package api

import (
	"encoding/json"
	"errors" // Import errors package
	"log"
	"net/http"

	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"gorm.io/gorm"
)

type AddChannelRequest struct {
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func AddChannelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingChannel models.MonitoredChannel
	// First, check if a channel with this username already exists in the database
	result := db.DB.Where("username = ?", req.Username).First(&existingChannel)

	if result.Error == nil {
		// Channel found in database, check if active and update if necessary
		log.Printf("Channel %s already exists in DB (ID: %d).", req.Username, existingChannel.ID)

		if existingChannel.IsActive != req.IsActive {
			// Update is_active status
			if err := db.DB.Model(&existingChannel).Update("is_active", req.IsActive).Error; err != nil {
				log.Printf("Failed to update is_active status for channel %s: %v", req.Username, err)
				http.Error(w, "Failed to update channel status", http.StatusInternalServerError)
				return
			}
			log.Printf("Updated is_active status for channel %s to %t", req.Username, req.IsActive)

			// If becoming active, start monitoring
			if req.IsActive {
				go monitor.StartMonitoringChannel(&existingChannel)
			}
		} else {
			log.Printf("Channel %s already exists and is_active status is the same.", req.Username)
			// If already active, ensure monitoring is running (could add a check here if needed)
			// For now, we assume the main startup loop handles starting monitors for active channels.
		}

		w.WriteHeader(http.StatusOK) // Use 200 OK for update/existing
		json.NewEncoder(w).Encode(existingChannel)
		return
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// An actual database error occurred (not just record not found)
		log.Printf("Database error checking for existing channel %s: %v", req.Username, result.Error)
		http.Error(w, "Database error checking channel", http.StatusInternalServerError)
		return
	}

	// Channel not found in database, proceed to fetch data from API
	log.Printf("Channel %s not found in DB. Fetching data from API.", req.Username)
	kickData, err := monitor.FetchChannelData(req.Username)
	if err != nil {
		log.Printf("Error fetching channel data for %s: %v", req.Username, err)
		http.Error(w, "Failed to fetch channel data", http.StatusInternalServerError)
		return
	}

	channel := models.MonitoredChannel{
		ID:         uint(kickData.ID), // Use the ID from the API
		ChatRoomID: uint(kickData.Chatroom.ID),
		Username:   req.Username,
		IsActive:   req.IsActive,
	}

	// Check again for potential race condition if another request added the channel
	// between the initial check and fetching data. This is less likely but good practice.
	var potentialExistingChannel models.MonitoredChannel
	if err := db.DB.First(&potentialExistingChannel, channel.ID).Error; err == nil {
		log.Printf("Race condition detected: Channel %s (ID: %d) was added by another process.", req.Username, channel.ID)
		http.Error(w, "Channel was added concurrently", http.StatusConflict)
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Database error checking for concurrent channel add for %s: %v", req.Username, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result = db.DB.Create(&channel)
	if result.Error != nil {
		log.Printf("Failed to add new channel %s to database: %v", req.Username, result.Error)
		http.Error(w, "Failed to add channel to database", http.StatusInternalServerError)
		return
	}

	log.Printf("Added new channel %s with ID %d to database", channel.Username, channel.ID)

	if channel.IsActive {
		go monitor.StartMonitoringChannel(&channel) // Start monitoring in a Go routine
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}
