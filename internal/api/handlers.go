package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

type HealthCheckResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func HealthCheckHandler(c echo.Context) error {
	response := HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "kick-monitor is alive",
	}
	return c.JSON(http.StatusOK, response)
}

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

	log.Printf("Received request to process lr for livestream ID: %d", req.LivestreamID)

	go func(livestreamID uint) {
		err := monitor.GenerateLivestreamReport(livestreamID)
		if err != nil {
			log.Printf("Error generating livestream lr for %d: %v", livestreamID, err)
		} else {
			log.Printf("Successfully generated livestream lr for %d", livestreamID)
		}
	}(req.LivestreamID)

	return c.JSON(http.StatusAccepted, map[string]string{"status": "processing_started", "message": "Livestream lr generation initiated."})
}

func getFullReport(query *gorm.DB) ([]monitor.FullLivestreamReportForProfile, error) {
	var livestreamReports []models.LivestreamReport
	if err := query.Find(&livestreamReports).Error; err != nil {
		return nil, fmt.Errorf("failed to find livestream reports: %w", err)
	}

	if len(livestreamReports) == 0 {
		return []monitor.FullLivestreamReportForProfile{}, nil
	}

	fullReports := make([]monitor.FullLivestreamReportForProfile, len(livestreamReports))
	for i, lr := range livestreamReports {
		fullReports[i].LivestreamReportRestructured = monitor.LivestreamReportRestructured{
			LivestreamID:          int(lr.LivestreamID),
			Title:                 lr.Title,
			ReportStartTime:       lr.ReportStartTime,
			DurationMinutes:       lr.DurationMinutes,
			AverageViewers:        lr.AverageViewers,
			PeakViewers:           lr.PeakViewers,
			LowestViewers:         lr.LowestViewers,
			Engagement:            lr.Engagement,
			TotalMessages:         lr.TotalMessages,
			HoursWatched:          lr.HoursWatched,
			UniqueChatters:        lr.UniqueChatters,
			MessagesFromApps:      lr.MessagesFromApps,
			ViewerCountsTimeline:  lr.ViewerCountsTimeline,
			MessageCountsTimeline: lr.MessageCountsTimeline,
			CreatedAt:             lr.CreatedAt,
		}
		// fmt.Println(i, lr)
		if lr.SpamReportID != nil {
			var spamReport models.SpamReport
			if err := db.DB.Where("id = ?", lr.SpamReportID).First(&spamReport).Error; err != nil {
				log.Printf("Warning: Failed to fetch spam report  %s for livestream id %s: %v", lr.SpamReportID.String(), lr.ID.String(), err)

			} else {
				fullReports[i].SpamReport = monitor.SpamReportRestructured{
					MessagesWithEmotes:         spamReport.MessagesWithEmotes,
					MessagesMultipleEmotesOnly: spamReport.MessagesWithEmotes,
					DuplicateMessagesCount:     spamReport.DuplicateMessagesCount,
					RepetitivePhrasesCount:     spamReport.RepetitivePhrasesCount,
					ExactDuplicateBursts:       spamReport.ExactDuplicateBursts,
					SimilarMessageBursts:       spamReport.SimilarMessageBursts,
					SuspiciousChatters:         spamReport.SuspiciousChatters,
				}
			}
		}
	}
	return fullReports, nil
}

// getLatestLivestreams handles the GET /livestreams/latest endpoint
func GetLatestLivestreams(c echo.Context) error {
	var latestLivestreams []models.LivestreamData

	// Using the Window Function approach (recommended for PostgreSQL)
	// This is typically the most efficient for this type of query.
	windowSQL := `
		SELECT *
		FROM (
			SELECT
				*,
				ROW_NUMBER() OVER (PARTITION BY livestream_id ORDER BY created_at DESC) as rn
			FROM
				livestream_data
		) AS subquery
		WHERE rn = 1
		ORDER BY livestream_id, created_at DESC;
	`
	err := db.DB.Raw(windowSQL).Scan(&latestLivestreams).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to get latest livestreams: %v", err)})
	}

	/*
		subQuery := db.DB.Model(&LivestreamData{}).
			Select("livestream_id, MAX(created_at) as created_at").
			Group("livestream_id")

		err = db.DB.Table("livestream_data").
			Joins("INNER JOIN (?) as t2 ON livestream_data.livestream_id = t2.livestream_id AND livestream_data.created_at = t2.created_at", subQuery).
			Find(&latestLivestreams).Error

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to get latest livestreams: %v", err)})
		}
	*/

	return c.JSON(http.StatusOK, latestLivestreams)
}

// getLatestLivestreamsByUsername handles the GET /livestreams/:username endpoint
func GetLatestLivestreamsByUsername(c echo.Context) error {
	username := c.Param("username") // Get username from URL path

	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username cannot be empty"})
	}

	// Step 1: Query MonitoredChannel to get ChannelID from Username
	var monitoredChannel models.MonitoredChannel
	result := db.DB.Where("username = ?", username).First(&monitoredChannel)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("channel with username '%s' not found", username)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to query channel by username: %v", result.Error)})
	}

	channelID := monitoredChannel.ChannelID
	log.Printf("Found ChannelID %d for username '%s'", channelID, username)

	// Step 2: Fetch latest LivestreamData entries for the found ChannelID
	var latestLivestreams []models.LivestreamData

	windowSQL := `
		SELECT *
		FROM (
			SELECT
				*,
				ROW_NUMBER() OVER (PARTITION BY livestream_id ORDER BY created_at DESC) as rn
			FROM
				livestream_data
			WHERE
				channel_id = ? -- Filter by ChannelID here
		) AS subquery
		WHERE rn = 1
		ORDER BY livestream_id, created_at DESC;
	`
	err := db.DB.Raw(windowSQL, channelID).Scan(&latestLivestreams).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to get latest livestreams for channel %d: %v", channelID, err)})
	}

	if len(latestLivestreams) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("no livestream data found for channel with username '%s'", username)})
	}

	return c.JSON(http.StatusOK, latestLivestreams)
}

// GetReportByUUIDHandler now takes echo.Context
func GetReportByUUIDHandler(c echo.Context) error {
	reportUUIDStr := c.Param("reportUUID") // Use c.Param for path variables
	reportUUID, err := uuid.Parse(reportUUIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid lr UUID format"})
	}

	fullReports, err := getFullReport(db.DB.Where("id = ?", reportUUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch lr: %v", err)})
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

// GetReportsByLivestreamIDHandler handles GET /livestream/id
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

func GetMonitoredChannelsHandler(c echo.Context) error {
	var channels []models.MonitoredChannel
	if err := db.DB.Order("username ASC").Find(&channels).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to fetch channels: %v", err)})
	}

	return c.JSON(http.StatusOK, channels)
}

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

func GetStreamerProfileHandler(c echo.Context) error {
	username := c.Param("username")

	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Username is required in the path"})
	}

	apiProfile, err := monitor.GetStreamerProfile(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("Streamer profile not found for username '%s'", username)})
		}
		log.Printf("Error fetcheing streamer profile for username '%s': %v", username, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Failed to build streamer profile: %v", err)})
	}

	return c.JSON(http.StatusOK, apiProfile)
}
