package api

import (
	"encoding/json"
	"errors" // Import errors package
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"github.com/google/uuid" // For parsing UUIDs
	"gorm.io/gorm"
)

type AddChannelRequest struct {
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type ProcessLivestreamReportRequest struct {
	LivestreamID uint `json:"livestream_id"`
}

type FullLivestreamReport struct {
	models.LivestreamReport
	SpamReport *models.SpamReport `json:"spam_report,omitempty"` // Embed SpamReport or make it nullable
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
		log.Printf("Channel %s already exists in DB (ID: %d).", req.Username, existingChannel.ChannelID)

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
		ChannelID:  uint(kickData.ID), // Use the ID from the API
		ChatroomID: uint(kickData.Chatroom.ID),
		Username:   req.Username,
		IsActive:   req.IsActive,
	}

	// Check again for potential race condition if another request added the channel
	// between the initial check and fetching data. This is less likely but good practice.
	var potentialExistingChannel models.MonitoredChannel
	if err := db.DB.First(&potentialExistingChannel, channel.ChannelID).Error; err == nil {
		log.Printf("Race condition detected: Channel %s (ID: %d) was added by another process.", req.Username, channel.ChannelID)
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

	log.Printf("Added new channel %s with ID %d to database", channel.Username, channel.ChannelID)

	if channel.IsActive {
		go monitor.StartMonitoringChannel(&channel) // Start monitoring in a Go routine
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

func ProcessLivestreamReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProcessLivestreamReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.LivestreamID == 0 {
		http.Error(w, "livestream_id is required and must be a valid ID", http.StatusBadRequest)
		return
	}

	log.Printf("Received request to process report for livestream ID: %d", req.LivestreamID)

	// In a real production app, you might want a queue here (e.g., RabbitMQ, Kafka)
	// to handle report generation requests, rather than spawning directly.
	// For now, we'll spawn a Go routine directly.
	go func(livestreamID uint) {
		err := monitor.GenerateLivestreamReport(livestreamID)
		if err != nil {
			log.Printf("Error generating livestream report for %d: %v", livestreamID, err)
			// In a real app, notify via metrics, logs, or an error service
		} else {
			log.Printf("Successfully generated livestream report for %d", livestreamID)
		}
	}(req.LivestreamID)

	w.WriteHeader(http.StatusAccepted) // 202 Accepted, processing is in progress
	json.NewEncoder(w).Encode(map[string]string{"status": "processing_started", "message": "Livestream report generation initiated."})
}

// getFullReport fetches a LivestreamReport and its associated SpamReport
func getFullReport(query *gorm.DB) ([]FullLivestreamReport, error) {
	var livestreamReports []models.LivestreamReport
	if err := query.Find(&livestreamReports).Error; err != nil {
		return nil, fmt.Errorf("failed to find livestream reports: %w", err)
	}

	if len(livestreamReports) == 0 {
		return []FullLivestreamReport{}, nil // Return empty slice if no reports found
	}

	fullReports := make([]FullLivestreamReport, len(livestreamReports))
	for i, lr := range livestreamReports {
		fullReports[i].LivestreamReport = lr
		if lr.SpamReportID != nil {
			var spamReport models.SpamReport
			if err := db.DB.Where("id = ?", lr.SpamReportID).First(&spamReport).Error; err != nil {
				log.Printf("Warning: Failed to fetch spam report %s for livestream report %s: %v", lr.SpamReportID.String(), lr.ID.String(), err)
				fullReports[i].SpamReport = nil // Set to nil if not found
			} else {
				fullReports[i].SpamReport = &spamReport
			}
		}
	}
	return fullReports, nil
}

// GetReportByUUIDHandler handles GET /reports/{report_uuid}
func GetReportByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	reportUUIDStr := r.URL.Path[len("/reports/"):]
	reportUUID, err := uuid.Parse(reportUUIDStr)
	if err != nil {
		http.Error(w, "Invalid report UUID format", http.StatusBadRequest)
		return
	}

	fullReports, err := getFullReport(db.DB.Where("id = ?", reportUUID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch report: %v", err), http.StatusInternalServerError)
		return
	}

	if len(fullReports) == 0 {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullReports[0]) // Return single report
}

// GetReportsByChannelIDHandler handles GET /channels/{channel_id}/reports
func GetReportsByChannelIDHandler(w http.ResponseWriter, r *http.Request) {
	channelIDStr := r.URL.Path[len("/channels/"):strings.LastIndex(r.URL.Path, "/reports")]
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid channel ID format", http.StatusBadRequest)
		return
	}

	fullReports, err := getFullReport(db.DB.Where("channel_id = ?", channelID).Order("report_start_time DESC"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch reports: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullReports)
}

// GetReportsByLivestreamIDHandler handles GET /livestreams/{livestream_id}/reports
func GetReportsByLivestreamIDHandler(w http.ResponseWriter, r *http.Request) {
	livestreamIDStr := r.URL.Path[len("/livestreams/"):strings.LastIndex(r.URL.Path, "/reports")]
	livestreamID, err := strconv.ParseUint(livestreamIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid livestream ID format", http.StatusBadRequest)
		return
	}

	fullReports, err := getFullReport(db.DB.Where("livestream_id = ?", livestreamID).Order("report_start_time DESC"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch reports: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullReports)
}

// GetMonitoredChannelsHandler handles GET /channels
func GetMonitoredChannelsHandler(w http.ResponseWriter, r *http.Request) {
	var channels []models.MonitoredChannel
	if err := db.DB.Order("username ASC").Find(&channels).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch channels: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

// GetChannelInfoHandler handles GET /channels/{channel_id}/info (latest snapshot)
func GetChannelInfoHandler(w http.ResponseWriter, r *http.Request) {
	channelIDStr := r.URL.Path[len("/channels/") : len(r.URL.Path)-len("/info")] // Careful with path parsing
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid channel ID format", http.StatusBadRequest)
		return
	}

	var latestChannelData models.ChannelData
	// Get the latest channel data for this channel_id
	if err := db.DB.Where("channel_id = ?", channelID).
		Order("created_at DESC").
		First(&latestChannelData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Channel info not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to fetch channel info: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestChannelData)
}

// // GetStreamerProfileHandler now fetches the pre-processed profile from the DB
// func GetStreamerProfileHandler(w http.ResponseWriter, r *http.Request) {
// 	channelIDStr := r.URL.Path[len("/channels/"):strings.LastIndex(r.URL.Path, "/profile")]
// 	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid channel ID format", http.StatusBadRequest)
// 		return
// 	}
//
// 	var profile models.StreamerProfile
// 	if err := db.DB.Where("channel_id = ?", channelID).First(&profile).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			http.Error(w, "Streamer profile not found (data might not have been collected yet)", http.StatusNotFound)
// 		} else {
// 			log.Printf("Error fetching streamer profile for channel %d: %v", channelID, err)
// 			http.Error(w, fmt.Sprintf("Failed to fetch streamer profile: %v", err), http.StatusInternalServerError)
// 		}
// 		return
// 	}
//
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(profile)
// }
