package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo
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
	SpamReport *models.SpamReport `json:"spam_report,omitempty"`
}

// AddChannelHandler now takes echo.Context
func AddChannelHandler(c echo.Context) error {
	req := new(AddChannelRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	var existingChannel models.MonitoredChannel
	result := db.DB.Where("username = ?", req.Username).First(&existingChannel)

	if result.Error == nil {
		log.Printf("Channel %s already exists in DB (ID: %d).", req.Username, existingChannel.ChannelID)

		if existingChannel.IsActive != req.IsActive {
			if err := db.DB.Model(&existingChannel).Update("is_active", req.IsActive).Error; err != nil {
				log.Printf("Failed to update is_active status for channel %s: %v", req.Username, err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update channel status"})
			}
			log.Printf("Updated is_active status for channel %s to %t", req.Username, req.IsActive)

			if req.IsActive {
				// go monitor.ProcessChannelData(existingChannel)
				go monitor.StartMonitoringChannel(&existingChannel)

			}
		} else {
			log.Printf("Channel %s already exists and is_active status is the same.", req.Username)
		}

		return c.JSON(http.StatusOK, existingChannel)
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Database error checking for existing channel %s: %v", req.Username, result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error checking channel"})
	}

	log.Printf("Channel %s not found in DB. Fetching data from API.", req.Username)
	kickData, err := monitor.FetchChannelData(req.Username)
	if err != nil {
		log.Printf("Error fetching channel data for %s: %v", req.Username, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch channel data"})
	}

	channel := models.MonitoredChannel{
		ChannelID:  uint(kickData.ID),
		ChatroomID: uint(kickData.Chatroom.ID),
		Username:   req.Username,
		IsActive:   req.IsActive,
	}

	var potentialExistingChannel models.MonitoredChannel
	if err := db.DB.First(&potentialExistingChannel, channel.ChannelID).Error; err == nil {
		log.Printf("Race condition detected: Channel %s (ID: %d) was added by another process.", req.Username, channel.ChannelID)
		return c.JSON(http.StatusConflict, map[string]string{"message": "Channel was added concurrently"})
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Database error checking for concurrent channel add for %s: %v", req.Username, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
	}

	result = db.DB.Create(&channel)
	if result.Error != nil {
		log.Printf("Failed to add new channel %s to database: %v", req.Username, result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to add channel to database"})
	}

	log.Printf("Added new channel %s with ID %d to database", channel.Username, channel.ChannelID)

	if channel.IsActive {
		go monitor.StartMonitoringChannel(&channel)
	}

	return c.JSON(http.StatusCreated, channel)
}

// ProcessLivestreamReportHandler now takes echo.Context
func ProcessLivestreamReportHandler(c echo.Context) error {
	req := new(ProcessLivestreamReportRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	if req.LivestreamID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "livestream_id is required and must be a valid ID"})
	}

	log.Printf("Received request to process report for livestream ID: %d", req.LivestreamID)

	go func(livestreamID uint) {
		err := monitor.GenerateLivestreamReport(livestreamID)
		if err != nil {
			log.Printf("Error generating livestream report for %d: %v", livestreamID, err)
		} else {
			log.Printf("Successfully generated livestream report for %d", livestreamID)
		}
	}(req.LivestreamID)

	return c.JSON(http.StatusAccepted, map[string]string{"status": "processing_started", "message": "Livestream report generation initiated."})
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

// GetReportByUUIDHandler now takes echo.Context
func GetReportByUUIDHandler(c echo.Context) error {
	reportUUIDStr := c.Param("reportUUID") // Use c.Param for path variables
	reportUUID, err := uuid.Parse(reportUUIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid report UUID format"})
	}

	fullReports, err := getFullReport(db.DB.Where("id = ?", reportUUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch report: %v", err)})
	}

	if len(fullReports) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Report not found"})
	}

	return c.JSON(http.StatusOK, fullReports[0])
}

// GetReportsByChannelIDHandler handles GET /channels/{channel_id}/reports
func GetReportsByChannelIDHandler(c echo.Context) error {
	channelIDStr := c.Param("channelID") // Use c.Param for path variables
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid channel ID format"})
	}

	fullReports, err := getFullReport(db.DB.Where("channel_id = ?", channelID).Order("report_start_time DESC"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch reports: %v", err)})
	}

	return c.JSON(http.StatusOK, fullReports)
}

// GetReportsByLivestreamIDHandler handles GET /livestreams/{livestream_id}/reports
func GetReportsByLivestreamIDHandler(c echo.Context) error {
	livestreamIDStr := c.Param("livestreamID") // Use c.Param for path variables
	livestreamID, err := strconv.ParseUint(livestreamIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid livestream ID format"})
	}

	fullReports, err := getFullReport(db.DB.Where("livestream_id = ?", livestreamID).Order("report_start_time DESC"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch reports: %v", err)})
	}

	return c.JSON(http.StatusOK, fullReports)
}

// GetMonitoredChannelsHandler handles GET /channels
func GetMonitoredChannelsHandler(c echo.Context) error {
	var channels []models.MonitoredChannel
	if err := db.DB.Order("username ASC").Find(&channels).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch channels: %v", err)})
	}

	return c.JSON(http.StatusOK, channels)
}

// GetChannelInfoHandler handles GET /channels/{channel_id}/info (latest snapshot)
func GetChannelInfoHandler(c echo.Context) error {
	channelIDStr := c.Param("channelID") // Use c.Param for path variables
	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid channel ID format"})
	}

	var latestChannelData models.ChannelData
	if err := db.DB.Where("channel_id = ?", channelID).
		Order("created_at DESC").
		First(&latestChannelData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Channel info not found"})
		} else {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch channel info: %v", err)})
		}
	}

	return c.JSON(http.StatusOK, latestChannelData)
}

// GetStreamerProfileHandler now takes echo.Context
// func GetStreamerProfileHandler(c echo.Context) error {
// 	channelIDStr := c.Param("channelID") // Use c.Param for path variables
// 	channelID, err := strconv.ParseUint(channelIDStr, 10, 64)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid channel ID format"})
// 	}
//
// 	var dbProfile models.StreamerProfile
// 	if err := db.DB.Where("channel_id = ?", channelID).First(&dbProfile).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return c.JSON(http.StatusNotFound, map[string]string{"message": "Streamer profile not found (data might not have been collected yet)"})
// 		} else {
// 			log.Printf("Error fetching streamer profile for channel %d: %v", channelID, err)
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch streamer profile: %v", err)})
// 		}
// 	}
//
// 	// Convert DB model (with []byte JSONB fields) to API model (with native Go types)
// 	apiProfile, err := monitor.ConvertDBProfileToAPIProfile(dbProfile)
// 	if err != nil {
// 		log.Printf("Error converting DB profile to API profile for channel %d: %v", channelID, err)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to prepare streamer profile for response"})
// 	}
//
// 	return c.JSON(http.StatusOK, apiProfile)
// }
