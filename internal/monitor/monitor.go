package monitor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	FetchInterval = 2 * time.Minute
	WebSocketURL  = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679" // Base WebSocket URL

	// Leeway for considering livestream data current
	LivestreamFreshnessLeeway = 20 * time.Second // 2 minutes + 20 seconds
	ReportTimeBlock           = 2 * time.Minute  // Viewer count timeline interval

	MessageTimelineBlock = 10 * time.Minute // Message count timeline interval

	// Spam Detection Thresholds (Adjust these values based on testing)
	ExactDuplicateBurstWindow   = 5 * time.Second // Time window for exact duplicate bursts
	ExactDuplicateBurstMinCount = 3               // Min identical messages in window for a burst

	SimilarMessageBurstWindow   = 10 * time.Second // Time window for similar message bursts
	SimilarMessageBurstMinCount = 4                // Min similar messages in window for a burst
	SimilarMessageMinSimilarity = 0.7              // Jaccard similarity threshold for "similar"

	RapidMessageBurstWindow   = 3 * time.Second // Time window for rapid messages by a user
	RapidMessageBurstMinCount = 5               // Min messages by same user in window for rapid burst
)

var ProxyURL string

// Structs for proxy response and Kick API data
type ProxyResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		URL       string            `json:"url"`
		Status    int               `json:"status"`
		Cookies   []any             `json:"cookies"`
		UserAgent string            `json:"userAgent"`
		Headers   map[string]string `json:"headers"`
		Response  string            `json:"response"` // HTML content
	} `json:"solution"`
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Version        string `json:"version"`
}

type KickChannelResponse struct {
	ID                  int    `json:"id"`
	UserID              int    `json:"user_id"`
	Slug                string `json:"slug"`
	IsBanned            bool   `json:"is_banned"`
	PlaybackURL         string `json:"playback_url"`
	VodEnabled          bool   `json:"vod_enabled"`
	SubscriptionEnabled bool   `json:"subscription_enabled"`
	IsAffiliate         bool   `json:"is_affiliate"`
	FollowersCount      int    `json:"followers_count"`
	SubscriberBadges    []any  `json:"subscriber_badges"` // Or define a proper struct
	BannerImage         struct {
		URL string `json:"url"`
	} `json:"banner_image"`
	Livestream         *KickLivestream `json:"livestream"` // Pointer to handle null
	Role               any             `json:"role"`       // Or define a proper struct
	Muted              bool            `json:"muted"`
	FollowerBadges     []any           `json:"follower_badges"`      // Or define a proper struct
	OfflineBannerImage any             `json:"offline_banner_image"` // Or define a proper struct
	Verified           bool            `json:"verified"`
	RecentCategories   []any           `json:"recent_categories"` // Or define a proper struct
	CanHost            bool            `json:"can_host"`
	User               *User           `json:"user"`
	Chatroom           *KickChatroom   `json:"chatroom"` // Pointer to handle null
}
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	AgreedToTerms   bool      `json:"agreed_to_terms"`
	EmailVerifiedAt time.Time `json:"email_verified_at"`
	Bio             string    `json:"bio"`
	Country         any       `json:"country"`
	State           any       `json:"state"`
	City            any       `json:"city"`
	Instagram       string    `json:"instagram"`
	Twitter         string    `json:"twitter"`
	Youtube         string    `json:"youtube"`
	Discord         string    `json:"discord"`
	Tiktok          string    `json:"tiktok"`
	Facebook        string    `json:"facebook"`
	ProfilePic      string    `json:"profile_pic"`
}

type KickLivestream struct {
	ID            int    `json:"id"`
	Slug          string `json:"slug"`
	ChannelID     int    `json:"channel_id"`
	CreatedAt     string `json:"created_at"` // Still string for unmarshalling
	SessionTitle  string `json:"session_title"`
	IsLive        bool   `json:"is_live"`
	RiskLevelID   any    `json:"risk_level_id"`  // Or define a proper struct
	StartTime     string `json:"start_time"`     // Still string for unmarshalling
	Source        any    `json:"source"`         // Or define a proper struct
	TwitchChannel any    `json:"twitch_channel"` // Or define a proper struct
	Duration      int    `json:"duration"`
	Language      string `json:"language"`
	IsMature      bool   `json:"is_mature"`
	ViewerCount   int    `json:"viewer_count"`
	Thumbnail     struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	LangISO    string          `json:"lang_iso"`
	Tags       json.RawMessage `json:"tags"`       // Use json.RawMessage to keep raw JSON for tags
	Categories []any           `json:"categories"` // Or define a proper struct
}

type KickChatroom struct {
	ID                   int    `json:"id"`
	ChatableType         string `json:"chatable_type"`
	ChannelID            int    `json:"channel_id"`
	CreatedAt            string `json:"created_at"` // Parse into time.Time
	UpdatedAt            string `json:"updated_at"` // Parse into time.Time
	ChatModeOld          string `json:"chat_mode_old"`
	ChatMode             string `json:"chat_mode"`
	SlowMode             bool   `json:"slow_mode"`
	ChatableID           int    `json:"chatable_id"`
	FollowersMode        bool   `json:"followers_mode"`
	SubscribersMode      bool   `json:"subscribers_mode"`
	EmotesMode           bool   `json:"emotes_mode"`
	MessageInterval      int    `json:"message_interval"`
	FollowingMinDuration int    `json:"following_min_duration"`
}

type ProxyRequestPayload struct {
	Cmd        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

// Struct to represent the generic WebSocket message structure
type IncomingMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

// Struct to store the latest livestream information
type LatestLivestreamInfo struct {
	LivestreamID uint
	FetchTime    time.Time
	IsLive       bool
}
type ChatMessageEventData struct {
	ID         string `json:"id"`
	ChatroomID int    `json:"chatroom_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"` // Original message send time
	Sender     struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Slug     string `json:"slug"`
		Identity struct {
			Color  string `json:"color"`
			Badges []any  `json:"badges"`
		} `json:"identity"`
	} `json:"sender"`
	Metadata json.RawMessage `json:"metadata"` // Use json.RawMessage for metadata
}

// this is a list of known chatbots/chat apps
var AppSenders = map[string]struct{}{
	"botrix":    {},
	"@fossabot": {},
	"fossabot":  {},
	"kicbot":    {},
}
var latestLivestream sync.Map // map[uint]LatestLivestreamInfo

var emoteRegex = regexp.MustCompile(`\[emote:\d+:\w+\]`)
var onlyEmotesRegex = regexp.MustCompile(`^(\s*\[emote:\d+:\w+\]\s*)+$`)
var suspiciousUsernameChecker = regexp.MustCompile(`(?i)(?:` +
	`bot|spam|ad|free\s*vbucks|nude\s*link|crypto|` + // Existing keywords
	`follow|sub|cash|giveaway|win|join|discord|telegram|link|onlyfans|of|` + // Added common keywords
	`\d{5,}$` + // More than 5 numbers at the end
	`)`)

// ReportMetrics holds all aggregated data during processing
type ReportMetrics struct {
	sync.Mutex

	TotalMessages              int
	UniqueChatters             map[string]struct{} // Using struct{} for a set (username)
	MessagesWithEmotes         int
	MessagesFromApps           int
	MessagesMultipleEmotesOnly int

	ExactDuplicateContents map[string]int              // content -> count for overall exact duplicates (for final count)
	ExactDuplicateBursts   []ExactDuplicateBurstReport // Identified bursts (slice)
	SimilarMessageBursts   []SimilarMessageBurstReport // Identified similar bursts (slice)
	RepetitivePhraseCounts map[string]int              // Phrase -> count (placeholder)
	SuspiciousChattersMap  map[int]struct{}            // map[SenderID]struct{} to track unique suspicious users by ID
	SuspiciousChattersList []SuspiciousChatterReport   // List of detailed reports for suspicious chatters (slice)

	ViewerCountsTimeline  []ViewerCountPoint
	MessageCountsTimeline []MessageCountPoint
}

// ViewerCountPoint for the timeline JSONB
type ViewerCountPoint struct {
	Time  time.Time `json:"time"`
	Count int       `json:"count"`
}

// MessageCountPoint for the timeline JSONB
type MessageCountPoint struct {
	Time  time.Time `json:"time"`
	Count int       `json:"count"`
}

// ExactDuplicateBurstReport for spam_reports table
type ExactDuplicateBurstReport struct {
	Username   string      `json:"username"` // Sender Username (slug)
	Content    string      `json:"content"`
	Count      int         `json:"count"`
	Timestamps []time.Time `json:"timestamps"`
}

// SimilarMessageBurstReport for spam_reports table
type SimilarMessageBurstReport struct {
	Username   string      `json:"username"` // Sender Username (slug)
	Pattern    string      `json:"pattern"`  // Representative content/pattern
	Count      int         `json:"count"`
	Timestamps []time.Time `json:"timestamps"`
}

// SuspiciousChatterReport for spam_reports table
type SuspiciousChatterReport struct {
	UserID            int         `json:"user_id"`
	Username          string      `json:"username"`
	PotentialIssues   []string    `json:"potential_issues"`
	MessageTimestamps []time.Time `json:"message_timestamps"`
	ExampleMessages   []string    `json:"example_messages"`
}

// NewReportMetrics initializes a new ReportMetrics instance
func NewReportMetrics() *ReportMetrics {
	return &ReportMetrics{
		UniqueChatters:         make(map[string]struct{}),
		ExactDuplicateContents: make(map[string]int),
		RepetitivePhraseCounts: make(map[string]int),
		SuspiciousChattersMap:  make(map[int]struct{}),
		SuspiciousChattersList: []SuspiciousChatterReport{},
		ViewerCountsTimeline:   []ViewerCountPoint{},
		MessageCountsTimeline:  []MessageCountPoint{},
	}
}

type LivestreamReportRestructured struct {
	LivestreamID          int             `json:"LivestreamID"`
	ReportStartTime       time.Time       `json:"ReportStartTime"`
	ReportEndTime         time.Time       `json:"ReportEndTime"`
	DurationMinutes       int             `json:"DurationMinutes"`
	AverageViewers        int             `json:"AverageViewers"`
	PeakViewers           int             `json:"PeakViewers"`
	LowestViewers         int             `json:"LowestViewers"`
	Engagement            float64         `json:"Engagement"`
	TotalMessages         int             `json:"TotalMessages"`
	UniqueChatters        int             `json:"UniqueChatters"`
	MessagesFromApps      int             `json:"MessagesFromApps"`
	ViewerCountsTimeline  json.RawMessage `json:"ViewerCountsTimeline"`
	MessageCountsTimeline json.RawMessage `json:"MessageCountsTimeline"`
	CreatedAt             time.Time       `json:"CreatedAt"`
}

type FullLivestreamReportForProfile struct {
	LivestreamReportRestructured
	SpamReport SpamReportRestructured `json:"spam_report"`
}

type StreamerProfileAPI struct {
	ChannelID           uint                             `json:"channel_id"`
	Username            string                           `json:"username"`
	Verified            bool                             `json:"verified"`
	IsBanned            bool                             `json:"is_banned"`
	VodEnabled          bool                             `json:"vod_enabled"`
	IsAffiliate         bool                             `json:"is_affiliate"`
	SubscriptionEnabled bool                             `json:"subscription_enabled"`
	FollowersCount      []models.FollowersCountPoint     `json:"followers_count"`
	Livestreams         []FullLivestreamReportForProfile `json:"livestreams"`

	Bio        string `json:"bio,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	TikTok     string `json:"tiktok,omitempty"`
	Country    string `json:"country,omitempty"`
	Discord    string `json:"discord,omitempty"`
	Twitter    string `json:"twitter,omitempty"`
	YouTube    string `json:"youtube,omitempty"`
	Facebook   string `json:"facebook,omitempty"`
	Instagram  string `json:"instagram,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"`
}

type SpamReportRestructured struct {
	MessagesWithEmotes         int             `json:"MessagesWithEmotes"`
	MessagesMultipleEmotesOnly int             `json:"MessagesMultipleEmotesOnly"`
	DuplicateMessagesCount     int             `json:"DuplicateMessagesCount"`
	RepetitivePhrasesCount     int             `json:"RepetitivePhrasesCount"`
	ExactDuplicateBursts       json.RawMessage `json:"ExactDuplicateBursts"`
	SimilarMessageBursts       json.RawMessage `json:"SimilarMessageBursts"`
	SuspiciousChatters         json.RawMessage `json:"SuspiciousChatters"`
}

func SetProxyURL(url string) error {
	if url == "" {
		return fmt.Errorf("apiclient: provided ProxyURL cannot be empty")
	}
	ProxyURL = url
	return nil
}

// StartMonitoringChannel initiates the data fetching and WebSocket routines for a channel.
func StartMonitoringChannel(channel *models.MonitoredChannel) {
	log.Printf("Starting monitoring for channel: %s (ID: %d)", channel.Username, channel.ChannelID)
	latestLivestream.Store(channel.ChannelID, LatestLivestreamInfo{}) // Start with a zero value
	// Start data fetching Go routine (uses proxy)
	go fetchDataAndPersist(channel)

	// Start WebSocket monitoring Go routine (does NOT use proxy)
	go startWebSocketMonitor(channel)
}

func FetchChannelData(username string) (*KickChannelResponse, error) {
	log.Printf("Fetching data for channel: %s via proxy", username)
	apiURL := fmt.Sprintf("https://kick.com/api/v2/channels/%s", username)

	if ProxyURL == "" {
		return nil, fmt.Errorf("ProxyURL not configured.")
	}
	proxyReqPayload := ProxyRequestPayload{
		Cmd:        "request.get",
		URL:        apiURL,
		MaxTimeout: 60000, // 60 seconds
	}

	proxyReqBody, err := json.Marshal(proxyReqPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling proxy request payload: %w", err)
	}

	resp, err := http.Post(ProxyURL, "application/json", bytes.NewBuffer(proxyReqBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to proxy for %s: %w", username, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading proxy response body for %s: %w", username, err)
	}

	var proxyResp ProxyResponse
	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling proxy response for %s: %w", username, err)
	}

	if proxyResp.Status != "ok" {
		return nil, fmt.Errorf("proxy returned non-ok status for %s: %s", username, proxyResp.Message)
	}

	// Extract JSON from HTML response within the proxy's solution.response
	jsonString, err := util.ExtractJSONFromHTML(proxyResp.Solution.Response)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON from HTML for %s: %w", username, err)
	}

	var kickData KickChannelResponse
	if err := json.Unmarshal([]byte(jsonString), &kickData); err != nil {
		return nil, fmt.Errorf("error unmarshalling Kick channel data for %s: %w", username, err)
	}

	return &kickData, nil
}

// fetchDataAndPersist periodically fetches and persists channel and livestream data.
func fetchDataAndPersist(channel *models.MonitoredChannel) {
	ticker := time.NewTicker(FetchInterval)
	defer ticker.Stop()

	// Initial fetch when the routine starts
	processChannelData(channel)

	for range ticker.C {
		processChannelData(channel)
	}
}

// ProcessChannelData: fetches, prints, and persists channel and livestream data, AND updates StreamerProfile
func processChannelData(channel *models.MonitoredChannel) { // Takes MonitoredChannel by value
	// log.Printf("Processing data for channel: %s (ID: %d, ChatroomID : %d)", channel.Username, channel.ChannelID, channel.ChatroomID)
	apiURL := fmt.Sprintf("https://kick.com/api/v2/channels/%s", channel.Username)

	proxyReqPayload := ProxyRequestPayload{
		Cmd:        "request.get",
		URL:        apiURL,
		MaxTimeout: 60000,
	}
	proxyReqBody, err := json.Marshal(proxyReqPayload)
	if err != nil {
		log.Printf("Error marshalling proxy request payload for channel %s: %v", channel.Username, err)
		return
	}

	resp, err := http.Post(ProxyURL, "application/json", bytes.NewBuffer(proxyReqBody))
	if err != nil {
		log.Printf("Error sending request to proxy for %s: %v", channel.Username, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading proxy response body for %s: %v", channel.Username, err)
		return
	}

	var proxyResp ProxyResponse
	if err := json.Unmarshal(body, &proxyResp); err != nil {
		log.Printf("Error unmarshalling proxy response for %s: %v", channel.Username, err)
		return
	}

	if proxyResp.Status != "ok" {
		log.Printf("Proxy returned non-ok status for %s: %s", channel.Username, proxyResp.Message)
		return
	}

	jsonString, err := util.ExtractJSONFromHTML(proxyResp.Solution.Response)
	if err != nil {
		log.Printf("Error extracting JSON from HTML for %s: %v", channel.Username, err)
		return
	}

	var kickData KickChannelResponse
	if err := json.Unmarshal([]byte(jsonString), &kickData); err != nil {
		log.Printf("Error unmarshalling Kick channel data for %s: %v", channel.Username, err)
		return
	}

	log.Printf("Fetched Channel Data for %s (ID: %d, ChatroomID : %d):\n", channel.Username, channel.ChannelID, channel.ChatroomID) // Log raw JSON

	channelData := models.ChannelData{
		ID:        uuid.New(),
		ChannelID: channel.ChannelID,
		Data:      []byte(jsonString),
	}
	if err := db.DB.Create(&channelData).Error; err != nil {
		log.Printf("Error saving channel data for %s: %v", channel.Username, err)
	} else {
		log.Printf("Saved channel data for %s (Channel ID: %d, UUID: %s)", channel.Username, channel.ChannelID, channelData.ID.String())
	}

	// Persist livestream data if available and update in-memory latest livestream info
	if kickData.Livestream != nil && kickData.Livestream.IsLive {
		// Parse timestamps from the livestream data string fields
		livestreamCreatedAt, err := time.Parse("2006-01-02 15:04:05", kickData.Livestream.CreatedAt)
		if err != nil {
			log.Printf("Error parsing livestream created_at timestamp for %s: %v", channel.Username, err)
			// Handle error or set a zero time if parsing fails
		}
		startTime, err := time.Parse("2006-01-02 15:04:05", kickData.Livestream.StartTime)
		if err != nil {
			log.Printf("Error parsing livestream start_time timestamp for %s: %v", channel.Username, err)
		}

		tagsData := []byte{}
		if kickData.Livestream.Tags != nil {
			tagsData = kickData.Livestream.Tags
		}

		livestreamID := uint(kickData.Livestream.ID)

		livestreamData := models.LivestreamData{
			ChannelID:    channel.ChannelID,
			LivestreamID: livestreamID, // Use the livestream ID from the data

			Slug:                kickData.Livestream.Slug,
			Tags:                tagsData,
			IsLive:              kickData.Livestream.IsLive,
			Duration:            kickData.Livestream.Duration,
			LangISO:             kickData.Livestream.LangISO,
			LivestreamCreatedAt: livestreamCreatedAt,
			StartTime:           startTime,
			ViewerCount:         kickData.Livestream.ViewerCount,
			SessionTitle:        kickData.Livestream.SessionTitle,
		}
		if err := db.DB.Create(&livestreamData).Error; err != nil {
			log.Printf("Error saving livestream data for %s (Livestream ID: %d): %v", channel.Username, livestreamData.LivestreamID, err)
		} else {
			log.Printf("Saved livestream data for %s (Channel ID: %d, Livestream ID: %d)", channel.Username, channel.ChannelID, livestreamData.LivestreamID)

			// Update in-memory latest livestream info
			latestLivestream.Store(channel.ChannelID, LatestLivestreamInfo{
				LivestreamID: livestreamID,
				FetchTime:    time.Now(), // Use the current time when data was successfully fetched
				IsLive:       kickData.Livestream.IsLive,
			})
			log.Printf("Updated in-memory latest livestream for channel %s (ID: %d) to LivestreamID: %d", channel.Username, channel.ChannelID, livestreamID)
		}
	} else {
		log.Printf("No active livestream data for channel: %s (ID: %d). Clearing in-memory latest livestream info.", channel.Username, channel.ChannelID)
		latestLivestream.Store(channel.ChannelID, LatestLivestreamInfo{})
	}

	err = streamerProfileBuilder(channel, kickData)
	if err != nil {
		log.Printf("Error updating streamer profile for channel %s (ID: %d): %v", channel.Username, channel.ChannelID, err)
	}
}

func createWebSocket(chatroomId uint) (*websocket.Conn, error) {
	params := url.Values{}
	params.Add("protocol", "7")
	params.Add("client", "js")
	params.Add("version", "7.4.0")
	params.Add("flash", "false")

	fullURL := WebSocketURL + "?" + params.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}

	subscribe := map[string]any{
		"event": "pusher:subscribe",
		"data": map[string]string{
			"auth":    "",
			"channel": fmt.Sprintf("chatrooms.%d.v2", chatroomId),
		},
	}

	if err := conn.WriteJSON(subscribe); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to subscribe to channel chatrooms.%d.v2: %w", chatroomId, err)
	}

	return conn, nil
}

func startWebSocketMonitor(channel *models.MonitoredChannel) {
	for {
		conn, err := createWebSocket(channel.ChatroomID)
		if err != nil {
			log.Printf("WebSocket connection error for channel %s (ID: %d): %v. Retrying in 5 seconds...", channel.Username, channel.ChatroomID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("WebSocket connected and subscribed for channel: %s (ID: %d)", channel.Username, channel.ChatroomID)

		// Read messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error for channel %s (ID: %d): %v. Attempting to reconnect...", channel.Username, channel.ChatroomID, err)
				conn.Close() // Close connection
				break
			}
			handleWebSocketMessage(channel, message)
		}
		time.Sleep(1 * time.Second)
	}
}

func handleWebSocketMessage(channel *models.MonitoredChannel, rawMessage []byte) {
	var msg IncomingMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error unmarshalling basic WebSocket message for %s: %v, raw message: %s", channel.Username, err, rawMessage)
		return
	}

	var currentLivestreamID *uint = nil // Default to null
	if info, ok := latestLivestream.Load(channel.ChannelID); ok {
		livestreamInfo := info.(LatestLivestreamInfo)
		// Check if the latest livestream data is recent and indicates a live stream
		if livestreamInfo.IsLive && time.Since(livestreamInfo.FetchTime) <= FetchInterval+LivestreamFreshnessLeeway {
			currentLivestreamID = &livestreamInfo.LivestreamID // Assign the livestream ID
		}
	}

	switch msg.Event {
	case "pusher_internal:subscription_succeeded":
		log.Printf("✅ WebSocket subscription succeeded for channel: %s (ID: %d, ChatroomID : %d)", channel.Username, channel.ChannelID, channel.ChatroomID)

	case "App\\Events\\ChatMessageEvent":
		// Unmarshal the Data string (which is JSON) into the ChatMessageEventData struct
		var chatMsgData ChatMessageEventData
		if err := json.Unmarshal([]byte(msg.Data), &chatMsgData); err != nil {
			log.Printf("Error unmarshalling ChatMessageEvent Data string for %s: %v, Data string: %s", channel.Username, err, msg.Data)
			return
		}

		// Parse the message send time using the correct format
		messageSendTime, err := time.Parse("2006-01-02T15:04:05Z07:00", chatMsgData.CreatedAt) // Correct format string
		if err != nil {
			log.Printf("Error parsing chat message created_at timestamp for %s: %v, value: %s", channel.Username, err, chatMsgData.CreatedAt)
		}

		// Parse the message ID string into a UUID
		messageUUID, err := uuid.Parse(chatMsgData.ID)
		if err != nil {
			log.Printf("Error parsing chat message ID string into UUID for %s: %v, value: %s", channel.Username, err, chatMsgData.ID)
			return
		}

		// Persist the chat message data with extracted fields
		chatMessage := models.ChatMessage{
			ID:           messageUUID,
			ChatroomID:   uint(chatMsgData.ChatroomID),
			Event:        msg.Event,
			LivestreamID: currentLivestreamID,
			CreatedAt:    time.Now(),

			// Populate extracted fields
			SenderID:        chatMsgData.Sender.ID,
			SenderUsername:  chatMsgData.Sender.Slug,
			Message:         chatMsgData.Content,
			Metadata:        chatMsgData.Metadata,
			MessageSendTime: messageSendTime,
		}

		if err := db.DB.Create(&chatMessage).Error; err != nil {
			log.Printf("Error saving chat message for %s (Message ID: %s): %v",
				channel.Username, chatMessage.ID.String(), err)
		} else {
			// temp disabled so we don't clutter
			// MessagePreview(channel, &chatMessage, currentLivestreamID, chatMsgData)
		}

	default:
		log.Printf("📩 Unhandled WebSocket event for %s ", channel.Username)
	}
}

func MessagePreview(channel *models.MonitoredChannel, chatMessage *models.ChatMessage, currentLivestreamID *uint, chatMsgData ChatMessageEventData) {
	var livestreamIDStr string
	if currentLivestreamID == nil {
		livestreamIDStr = "NULL"
	} else {
		livestreamIDStr = strconv.FormatUint(uint64(*currentLivestreamID), 10)
	}

	log.Printf("Saved chat message for %s (LID: %s): %s",
		channel.Username,
		livestreamIDStr,
		chatMsgData.Content,
	)
}

func GenerateLivestreamReport(livestreamID uint) error {
	var monitoredChannel models.MonitoredChannel
	subQuery := db.DB.Model(&models.LivestreamData{}).Select("channel_id").Where("livestream_id = ?", livestreamID)
	err := db.DB.Where("channel_id IN (?)", subQuery).First(&monitoredChannel).Error
	if err != nil {
		return fmt.Errorf("failed to find channel for livestream %d: %w", livestreamID, err)
	}

	ChannelID := monitoredChannel.ChannelID
	channelUsername := monitoredChannel.Username

	var streamActualStartTime time.Time
	if err := db.DB.Model(&models.LivestreamData{}).
		Select("start_time").
		Where("livestream_id = ?", livestreamID).
		Order("start_time ASC").
		Limit(1).
		Scan(&streamActualStartTime).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Warning: No initial start_time found for livestream %d in livestream_data. Using min message time.", livestreamID)
		} else {
			return fmt.Errorf("failed to get actual stream start_time for livestream %d: %w", livestreamID, err)
		}
	}

	// Find the min/max message send times for this livestream to define the report window
	var minMessageTime time.Time
	var maxMessageTime time.Time

	row := db.DB.Model(&models.ChatMessage{}).
		Select("MIN(message_send_time), MAX(message_send_time)").
		Where("livestream_id = ?", livestreamID).
		Row()

	if err := row.Scan(&minMessageTime, &maxMessageTime); err != nil {
		if err == gorm.ErrRecordNotFound || minMessageTime.IsZero() {
			log.Printf("No chat messages found for livestream ID: %d in the specified time range. Report cannot be generated.", livestreamID)
			return fmt.Errorf("no chat messages for livestream %d", livestreamID)
		}
		return fmt.Errorf("failed to get message time range for livestream %d: %w", livestreamID, err)
	}

	reportStartTime := minMessageTime.Truncate(MessageTimelineBlock)
	reportEndTime := maxMessageTime.Add(MessageTimelineBlock).Truncate(MessageTimelineBlock)

	// If streamActualStartTime was not found or is later than minMessageTime, use minMessageTime
	if streamActualStartTime.IsZero() || streamActualStartTime.After(reportStartTime) {

		streamActualStartTime = reportStartTime
	}

	// Calculate duration in minutes (from reportStartTime to reportEndTime)
	durationMinutes := int(reportEndTime.Sub(reportStartTime).Minutes())

	// 2. Fetch all relevant chat messages for the livestream
	var chatMessages []models.ChatMessage
	if err := db.DB.Where("livestream_id = ?", livestreamID).
		Order("message_send_time ASC").
		Find(&chatMessages).Error; err != nil {
		return fmt.Errorf("failed to fetch chat messages for livestream %d: %w", livestreamID, err)
	}
	log.Printf("Fetched %d chat messages for livestream %d", len(chatMessages), livestreamID)

	// 3. Fetch all relevant viewer counts for the channel and time range
	var viewerCounts []models.LivestreamData
	if err := db.DB.Where("channel_id = ? AND created_at >= ? AND created_at <= ?",
		ChannelID, reportStartTime.Add(-ReportTimeBlock), reportEndTime.Add(ReportTimeBlock)).
		Order("created_at ASC").
		Find(&viewerCounts).Error; err != nil {
		return fmt.Errorf("failed to fetch viewer counts for channel %d: %w", ChannelID, err)
	}
	log.Printf("Fetched %d viewer count records for channel %d", len(viewerCounts), ChannelID)

	metrics := NewReportMetrics()

	messageProcessingChan := make(chan models.ChatMessage, len(chatMessages))
	for _, msg := range chatMessages {
		messageProcessingChan <- msg
	}
	close(messageProcessingChan)

	numWorkers := 4
	var wg sync.WaitGroup

	for range make([]struct{}, numWorkers) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range messageProcessingChan {
				processSingleMessage(msg, metrics)
			}
		}()
	}

	wg.Wait()

	metrics.TotalMessages = len(chatMessages)

	var viewerTimelineJSON []byte
	var messageTimelineJSON []byte

	metrics.ViewerCountsTimeline = buildViewerCountTimeline(viewerCounts, reportStartTime, reportEndTime)
	viewerTimelineJSON, err = json.Marshal(metrics.ViewerCountsTimeline) // Assign here
	if err != nil {
		log.Printf("Error marshalling viewer counts timeline for livestream %d: %v", livestreamID, err)
		viewerTimelineJSON = []byte("[]")
	}

	metrics.MessageCountsTimeline = buildMessageCountTimeline(chatMessages, reportStartTime, reportEndTime)
	messageTimelineJSON, err = json.Marshal(metrics.MessageCountsTimeline) // Assign here
	if err != nil {
		log.Printf("Error marshalling message counts timeline for livestream %d: %v", livestreamID, err)
		messageTimelineJSON = []byte("[]")
	}

	averageViewers, peakViewers, lowestViewers := calculateViewerAnalytics(viewerCounts)

	engagement := 0.0
	if averageViewers > 0 {
		engagement = (float64(len(metrics.UniqueChatters)) / float64(averageViewers)) * 100.0
	}

	// --- Spam Analysis - Post-processing after all messages have been individually processed ---
	userMessageHistory := make(map[int][]models.ChatMessage)
	for _, msg := range chatMessages {
		userMessageHistory[msg.SenderID] = append(userMessageHistory[msg.SenderID], msg)
	}

	for _, messages := range userMessageHistory {
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].MessageSendTime.Before(messages[j].MessageSendTime)
		})

		// Check for Exact Duplicate Bursts
		for i := 0; i < len(messages); i++ {
			currentMsg := messages[i]
			exactBurstCount := 1
			burstTimestamps := []time.Time{currentMsg.MessageSendTime}

			for j := i + 1; j < len(messages) && messages[j].MessageSendTime.Sub(currentMsg.MessageSendTime) <= ExactDuplicateBurstWindow; j++ {
				if util.NormalizeChatMessage(messages[j].Message) == util.NormalizeChatMessage(currentMsg.Message) {
					exactBurstCount++
					burstTimestamps = append(burstTimestamps, messages[j].MessageSendTime)
				}
			}

			if exactBurstCount >= ExactDuplicateBurstMinCount {
				metrics.Lock()
				metrics.ExactDuplicateBursts = append(metrics.ExactDuplicateBursts, ExactDuplicateBurstReport{
					Username:   currentMsg.SenderUsername,
					Content:    currentMsg.Message,
					Count:      exactBurstCount,
					Timestamps: util.UniqueSortedTimes(burstTimestamps),
				})
				metrics.Unlock()
				i += exactBurstCount - 1
			}
		}

		// Check for Similar Message Bursts
		for i := 0; i < len(messages); i++ {
			currentMsg := messages[i]
			similarMessagesInBurst := []string{currentMsg.Message}
			similarBurstCount := 1
			burstTimestamps := []time.Time{currentMsg.MessageSendTime}

			for j := i + 1; j < len(messages) && messages[j].MessageSendTime.Sub(currentMsg.MessageSendTime) <= SimilarMessageBurstWindow; j++ {
				if util.JaccardSimilarity(util.NormalizeChatMessage(currentMsg.Message), util.NormalizeChatMessage(messages[j].Message)) >= SimilarMessageMinSimilarity {
					similarMessagesInBurst = append(similarMessagesInBurst, messages[j].Message)
					similarBurstCount++
					burstTimestamps = append(burstTimestamps, messages[j].MessageSendTime)
				}
			}

			if similarBurstCount >= SimilarMessageBurstMinCount {
				metrics.Lock()
				metrics.SimilarMessageBursts = append(metrics.SimilarMessageBursts, SimilarMessageBurstReport{
					Username:   currentMsg.SenderUsername,
					Pattern:    strings.Join(util.UniqueStrings(similarMessagesInBurst), " / "),
					Count:      similarBurstCount,
					Timestamps: util.UniqueSortedTimes(burstTimestamps),
				})
				metrics.Unlock()
				i += similarBurstCount - 1
			}
		}

		for i := 0; i < len(messages); i++ {
			currentMsg := messages[i]
			rapidBurstCount := 1
			burstTimestamps := []time.Time{currentMsg.MessageSendTime}
			exampleMessages := []string{currentMsg.Message}

			for j := i + 1; j < len(messages) && messages[j].MessageSendTime.Sub(currentMsg.MessageSendTime) <= RapidMessageBurstWindow; j++ {
				rapidBurstCount++
				burstTimestamps = append(burstTimestamps, messages[j].MessageSendTime)
				exampleMessages = append(exampleMessages, messages[j].Message)
			}

			if rapidBurstCount >= RapidMessageBurstMinCount {
				metrics.Lock()
				if _, ok := metrics.SuspiciousChattersMap[currentMsg.SenderID]; !ok {
					metrics.SuspiciousChattersMap[currentMsg.SenderID] = struct{}{}
					metrics.SuspiciousChattersList = append(metrics.SuspiciousChattersList, SuspiciousChatterReport{
						UserID:            currentMsg.SenderID,
						Username:          currentMsg.SenderUsername,
						PotentialIssues:   []string{"rapid_message_bursts"},
						MessageTimestamps: util.UniqueSortedTimes(burstTimestamps),
						ExampleMessages:   util.UniqueStrings(exampleMessages),
					})
				} else {
					for k := range metrics.SuspiciousChattersList {
						if metrics.SuspiciousChattersList[k].UserID == currentMsg.SenderID {
							if !util.ContainsString(metrics.SuspiciousChattersList[k].PotentialIssues, "rapid_message_bursts") {
								metrics.SuspiciousChattersList[k].PotentialIssues = append(metrics.SuspiciousChattersList[k].PotentialIssues, "rapid_message_bursts")
							}
							metrics.SuspiciousChattersList[k].MessageTimestamps = util.UniqueSortedTimes(append(metrics.SuspiciousChattersList[k].MessageTimestamps, burstTimestamps...))
							metrics.SuspiciousChattersList[k].ExampleMessages = util.UniqueStrings(append(metrics.SuspiciousChattersList[k].ExampleMessages, exampleMessages...))
							break
						}
					}
				}
				metrics.Unlock()
				i += rapidBurstCount - 1
			}
		}
	}

	// Check for Suspicious Usernames
	for userID, msgs := range userMessageHistory {
		if len(msgs) > 0 {
			usernameToCheck := msgs[0].SenderUsername
			if suspiciousUsernameChecker.MatchString(usernameToCheck) {
				metrics.Lock()
				if _, ok := metrics.SuspiciousChattersMap[userID]; !ok {
					metrics.SuspiciousChattersMap[userID] = struct{}{}
					metrics.SuspiciousChattersList = append(metrics.SuspiciousChattersList, SuspiciousChatterReport{
						UserID:            userID,
						Username:          usernameToCheck,
						PotentialIssues:   []string{"suspicious_username"},
						MessageTimestamps: []time.Time{},
						ExampleMessages:   []string{},
					})
				} else {
					for k := range metrics.SuspiciousChattersList {
						if metrics.SuspiciousChattersList[k].UserID == userID {
							if !util.ContainsString(metrics.SuspiciousChattersList[k].PotentialIssues, "suspicious_username") {
								metrics.SuspiciousChattersList[k].PotentialIssues = append(metrics.SuspiciousChattersList[k].PotentialIssues, "suspicious_username")
							}
							break
						}
					}
				}
				metrics.Unlock()
			}
		}
	}

	// Sort bursts by count (higher count first)
	sort.Slice(metrics.ExactDuplicateBursts, func(i, j int) bool {
		return metrics.ExactDuplicateBursts[i].Count > metrics.ExactDuplicateBursts[j].Count
	})
	sort.Slice(metrics.SimilarMessageBursts, func(i, j int) bool {
		return metrics.SimilarMessageBursts[i].Count > metrics.SimilarMessageBursts[j].Count
	})

	// Create Spam Report							ID: string(report.ID),
	spamReport := models.SpamReport{
		ID:                 uuid.New(),
		LivestreamReportID: uuid.Nil, // Will be set after livestream report is created
		ChannelID:          ChannelID,
		LivestreamID:       livestreamID,
		CreatedAt:          time.Now(),
	}

	// Populate spam report fields
	totalExactDuplicates := 0
	metrics.Lock()
	for _, count := range metrics.ExactDuplicateContents {
		if count > 1 {
			totalExactDuplicates += (count - 1)
		}
	}
	metrics.Unlock()
	spamReport.DuplicateMessagesCount = totalExactDuplicates

	exactBurstsJSON, err := json.Marshal(metrics.ExactDuplicateBursts)
	if err != nil {
		log.Printf("Error marshalling exact duplicate bursts for spam report: %v", err)
		exactBurstsJSON = []byte("[]")
	}
	spamReport.ExactDuplicateBursts = exactBurstsJSON

	similarBurstsJSON, err := json.Marshal(metrics.SimilarMessageBursts)
	if err != nil {
		log.Printf("Error marshalling similar message bursts for spam report: %v", err)
		similarBurstsJSON = []byte("[]")
	}
	spamReport.SimilarMessageBursts = similarBurstsJSON

	suspiciousChattersJSON, err := json.Marshal(metrics.SuspiciousChattersList)
	if err != nil {
		log.Printf("Error marshalling suspicious chatters for spam report: %v", err)
		suspiciousChattersJSON = []byte("[]")
	}
	spamReport.SuspiciousChatters = suspiciousChattersJSON

	spamReport.RepetitivePhrasesCount = 0 // Placeholder

	// Moved emote counts to spam report
	spamReport.MessagesWithEmotes = metrics.MessagesWithEmotes
	spamReport.MessagesMultipleEmotesOnly = metrics.MessagesMultipleEmotesOnly

	if err := db.DB.Create(&spamReport).Error; err != nil {
		return fmt.Errorf("failed to save spam report for %d: %w", livestreamID, err)
	}
	log.Printf("Successfully generated spam report for livestream ID %d (Spam Report ID: %s)", livestreamID, spamReport.ID.String())

	var sessionTitle string
	err = db.DB.Model(&models.LivestreamData{}).Select("session_title").Where("livestream_id = ?", livestreamID).Order("created_at DESC").First(&sessionTitle).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("No entry found for LivestreamID: %d\n", livestreamID)
		} else {
			fmt.Printf("Error fetching only SessionTitle: %v\n", err)
		}
	} else {
		fmt.Printf("Session Title (only fetched) for LivestreamID %d (last entry): %s\n", livestreamID, sessionTitle)
	}

	hoursWatched := CalculateWatchHours(metrics.ViewerCountsTimeline)

	// Create Main Livestream Report
	report := models.LivestreamReport{
		ID:              uuid.New(),
		LivestreamID:    livestreamID,
		Title:           sessionTitle,
		ChannelID:       ChannelID,
		Username:        channelUsername,
		ReportStartTime: reportStartTime,
		ReportEndTime:   reportEndTime,
		DurationMinutes: durationMinutes,

		// Viewer Analytics
		AverageViewers:   averageViewers,
		PeakViewers:      peakViewers,
		LowestViewers:    lowestViewers,
		Engagement:       engagement,
		HoursWatched:     hoursWatched,
		TotalMessages:    metrics.TotalMessages,
		UniqueChatters:   len(metrics.UniqueChatters),
		MessagesFromApps: metrics.MessagesFromApps,

		SpamReportID: &spamReport.ID,

		ViewerCountsTimeline:  viewerTimelineJSON,
		MessageCountsTimeline: messageTimelineJSON,

		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&report).Error; err != nil {
		return fmt.Errorf("failed to save livestream report for %d: %w", livestreamID, err)
	}

	err = UpdateStreamerProfileLivestreams(ChannelID, report.ID)
	if err != nil {
		log.Printf("ERROR: Failed to update streamer profile with new report UUID %s for channel %d: %v", report.ID.String(), ChannelID, err)
	}

	spamReport.LivestreamReportID = report.ID
	if err := db.DB.Save(&spamReport).Error; err != nil {
		log.Printf("Warning: Failed to update spam_report %s with livestream_report_id %s: %v", spamReport.ID.String(), report.ID.String(), err)
	}

	log.Printf("Successfully generated main livestream report for livestream ID %d (Report ID: %s)", livestreamID, report.ID.String())
	return nil
}

func processSingleMessage(msg models.ChatMessage, metrics *ReportMetrics) {
	metrics.Lock() // Lock for general metric updates
	defer metrics.Unlock()

	// Unique Chatters
	metrics.UniqueChatters[msg.SenderUsername] = struct{}{}

	// Emote Detection
	if emoteRegex.MatchString(msg.Message) {
		metrics.MessagesWithEmotes++
		// Check for messages with *only* emotes
		normalizedMessage := strings.TrimSpace(msg.Message)
		if onlyEmotesRegex.MatchString(normalizedMessage) {
			metrics.MessagesMultipleEmotesOnly++
		}
	}

	// Messages from Apps
	if _, isApp := AppSenders[msg.SenderUsername]; isApp {
		metrics.MessagesFromApps++
	}

	normalizedContent := util.NormalizeChatMessage(msg.Message)
	metrics.ExactDuplicateContents[normalizedContent]++

	// For other spam detections (bursts, similar messages, rapid bursts),
	// we need context of other messages from the same user, which is done post-processing
	// after collecting all messages per user.
	// The `processSingleMessage` is for basic, independent message-level metrics.
	// More complex, sequence-dependent metrics are done in `GenerateLivestreamReport`.
}

func buildViewerCountTimeline(viewerCounts []models.LivestreamData, reportStartTime, reportEndTime time.Time) []ViewerCountPoint {
	timeline := []ViewerCountPoint{}
	if len(viewerCounts) == 0 {
		return timeline
	}

	currentBlockTime := reportStartTime.Truncate(ReportTimeBlock)

	for currentBlockTime.Before(reportEndTime) {
		blockEndTime := currentBlockTime.Add(ReportTimeBlock)

		var lastCountInBlock int
		foundInBlock := false

		for i := len(viewerCounts) - 1; i >= 0; i-- {
			vc := viewerCounts[i]
			if vc.CreatedAt.Before(blockEndTime) && !vc.CreatedAt.Before(currentBlockTime) {
				lastCountInBlock = vc.ViewerCount
				foundInBlock = true
				break
			}
		}

		if !foundInBlock && len(timeline) > 0 {
			lastCountInBlock = timeline[len(timeline)-1].Count
			foundInBlock = true
		} else if !foundInBlock {
			lastCountInBlock = 0
		}

		timeline = append(timeline, ViewerCountPoint{
			Time:  currentBlockTime,
			Count: lastCountInBlock,
		})

		currentBlockTime = blockEndTime
	}

	return timeline
}

func buildLivestreamsList(channel *models.MonitoredChannel) []uuid.UUID {
	var livestreamReports []uuid.UUID
	if err := db.DB.Model(&models.LivestreamReport{}).
		Select("id").
		Where("channel_id = ?", channel.ChannelID).
		Pluck("id", &livestreamReports).Error; err != nil {
		log.Printf("Failed to fetch livestream IDs for channel %d: %v", channel.ChannelID, err)
	}
	return livestreamReports
}

func buildMessageCountTimeline(messages []models.ChatMessage, reportStartTime, reportEndTime time.Time) []MessageCountPoint {
	timeline := []MessageCountPoint{}
	if len(messages) == 0 {
		return timeline
	}

	currentBlockTime := reportStartTime.Truncate(MessageTimelineBlock)

	blockCounts := make(map[time.Time]int)
	for _, msg := range messages {
		block := msg.MessageSendTime.Truncate(MessageTimelineBlock)
		blockCounts[block]++
	}

	for currentBlockTime.Before(reportEndTime) {
		count := blockCounts[currentBlockTime]
		timeline = append(timeline, MessageCountPoint{
			Time:  currentBlockTime,
			Count: count,
		})
		currentBlockTime = currentBlockTime.Add(MessageTimelineBlock)
	}

	return timeline
}

func calculateViewerAnalytics(viewerCounts []models.LivestreamData) (average, peak, lowest int) {
	if len(viewerCounts) == 0 {
		return 0, 0, 0
	}

	totalViewers := 0
	peak = 0
	lowest = math.MaxInt32 // Initialize lowest with a very high number

	for _, vc := range viewerCounts {
		totalViewers += vc.ViewerCount
		if vc.ViewerCount > peak {
			peak = vc.ViewerCount
		}
		if vc.ViewerCount < lowest {
			lowest = vc.ViewerCount
		}
	}

	average = totalViewers / len(viewerCounts)

	return average, peak, lowest
}

func CalculateWatchHours(points []ViewerCountPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	var totalSeconds float64
	for i := 1; i < len(points); i++ {
		dt := points[i].Time.Sub(points[i-1].Time).Seconds()
		totalSeconds += float64(points[i-1].Count) * dt
	}

	return totalSeconds / 3600.0
}

func streamerProfileBuilder(channel *models.MonitoredChannel, kickData KickChannelResponse) error {
	log.Println("Streamer profile builder Ran for:", channel.Username)
	var profile models.StreamerProfile
	var existingProfile models.StreamerProfile

	result := db.DB.Where("channel_id = ?", channel.ChannelID).First(&existingProfile)

	if result.Error == nil {
		profile = existingProfile // If exists, load existing data
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("database error checking existing profile for channel %d: %w", channel.ChannelID, result.Error)
	} else {
		profile.ChannelID = channel.ChannelID
	}

	// Populate common fields
	profile.Username = channel.Username
	profile.Verified = kickData.Verified
	profile.IsBanned = kickData.IsBanned
	profile.VodEnabled = kickData.VodEnabled
	profile.IsAffiliate = kickData.IsAffiliate
	profile.SubscriptionEnabled = kickData.SubscriptionEnabled

	// Populate fields from KickChannelResponse.User
	if kickData.User != nil {
		profile.Bio = kickData.User.Bio
		// Type assertions for any{} fields (Country, State, City)
		if countryStr, ok := kickData.User.Country.(string); ok {
			profile.Country = countryStr
		} else {
			profile.Country = ""
		}
		if stateStr, ok := kickData.User.State.(string); ok {
			profile.State = stateStr
		} else {
			profile.State = ""
		}
		if cityStr, ok := kickData.User.City.(string); ok {
			profile.City = cityStr
		} else {
			profile.City = ""
		}

		// Social media links (these are direct strings in your User struct)
		profile.TikTok = kickData.User.Tiktok
		profile.Discord = kickData.User.Discord
		profile.Twitter = kickData.User.Twitter
		profile.YouTube = kickData.User.Youtube
		profile.Facebook = kickData.User.Facebook
		profile.Instagram = kickData.User.Instagram
		profile.ProfilePic = kickData.User.ProfilePic
	} else {
		log.Printf("Warning: KickChannelResponse.User is nil for channel %d. Some profile fields will be empty.", channel.ChannelID)
		// Clear fields if User is nil and this is an update
		profile.Bio = ""
		profile.Country = ""
		profile.State = ""
		profile.City = ""
		profile.TikTok = ""
		profile.Discord = ""
		profile.Twitter = ""
		profile.YouTube = ""
		profile.Facebook = ""
		profile.Instagram = ""
		profile.ProfilePic = ""
	}

	// Build followers_count timeline from all historical channel_data
	var allChannelData []models.ChannelData
	if err := db.DB.Where("channel_id = ?", channel.ChannelID).Order("created_at ASC").Find(&allChannelData).Error; err != nil {
		log.Printf("Warning: Failed to fetch historical channel_data for followers timeline for channel %d: %v", channel.ChannelID, err)
		emptyJsonArray, err := json.Marshal([]models.FollowersCountPoint{})
		if err != nil {
			log.Fatalf("failed to create an empty array:%v", err)
		}
		profile.FollowersCount = emptyJsonArray
	} else {
		followersTimeline := make([]models.FollowersCountPoint, 0, len(allChannelData))
		for _, cd := range allChannelData {
			var historicalKickResponse KickChannelResponse
			if err := json.Unmarshal(cd.Data, &historicalKickResponse); err != nil {
				log.Printf("Warning: Error unmarshalling historical channel_data for followers timeline for channel %d (ID: %s): %v", channel.ChannelID, cd.ID.String(), err)
				continue
			}
			followersTimeline = append(followersTimeline, models.FollowersCountPoint{
				Time:  cd.CreatedAt,
				Count: historicalKickResponse.FollowersCount,
			})
		}

		followersTimelineJSON, err := json.Marshal(followersTimeline)
		if err != nil {
			log.Fatalf("Error: failed to marshal followersTimeline %v", err)
		}

		profile.FollowersCount = followersTimelineJSON // Assign directly, GORM handles JSONB
	}

	livestreamList := buildLivestreamsList(channel)
	livestreamListJSON, err := json.Marshal(livestreamList)
	if err != nil {
		log.Fatalf("Error: failed to marshal livestreamList %v", err)
	}

	// asign livestream list to profile
	profile.Livestreams = livestreamListJSON

	// Save/Update the StreamerProfile
	if result.Error == nil {
		if err := db.DB.Save(&profile).Error; err != nil {
			return fmt.Errorf("failed to update streamer profile for channel %d: %w", channel.ChannelID, err)
		}
		log.Printf("Updated streamer profile for channel %s (ID: %d)", channel.Username, channel.ChannelID)
	} else {
		if err := db.DB.Create(&profile).Error; err != nil {
			return fmt.Errorf("failed to create streamer profile for channel %d: %w", channel.ChannelID, err)
		}
		log.Printf("Created streamer profile for channel %s (ID: %d)", channel.Username, channel.ChannelID)
	}

	return nil
}

func UpdateStreamerProfileLivestreams(ChannelID uint, newReportUUID uuid.UUID) error {
	var profile models.StreamerProfile
	if err := db.DB.Where("channel_id = ?", ChannelID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Warning: Streamer profile not found for channel %d. Cannot update livestreams list. Creating empty profile.", ChannelID)
			return nil
		}
		return fmt.Errorf("failed to fetch streamer profile for channel %d to update livestreams: %w", ChannelID, err)
	}

	// Profile found, unmarshal current Livestreams from []byte to Go-native slice
	var nativeLivestreams []uuid.UUID
	if err := json.Unmarshal(profile.Livestreams, &nativeLivestreams); err != nil {
		return fmt.Errorf("failed to unmarshal existing Livestreams for channel %d: %w", ChannelID, err)
	}

	// Check if the UUID is already in the list to prevent duplicates
	if slices.Contains(nativeLivestreams, newReportUUID) {
		log.Printf("Livestream report UUID %s already exists in profile for channel %d. Not appending.", newReportUUID.String(), ChannelID)
		return nil
	}

	// add a new livestream uuid to previous list
	nativeLivestreams = append(nativeLivestreams, newReportUUID)

	nativeLivestreamsJSON, err := json.Marshal(nativeLivestreams)
	if err != nil {
		return fmt.Errorf("failed to marshal updated Livestreams for channel %d: %w", ChannelID, err)
	}

	// assign livestream list to profile
	profile.Livestreams = nativeLivestreamsJSON

	if err := db.DB.Save(&profile).Error; err != nil {
		return fmt.Errorf("failed to update streamer profile livestreams for channel %d: %w", ChannelID, err)
	}
	log.Printf("Added livestream report UUID %s to profile for channel %d", newReportUUID.String(), ChannelID)
	return nil
}

func GetStreamerProfile(username string) (StreamerProfileAPI, error) {
	var apiProfile StreamerProfileAPI

	var dbProfile models.StreamerProfile
	if err := db.DB.Where("username = ?", username).First(&dbProfile).Error; err != nil {
		return StreamerProfileAPI{}, fmt.Errorf("failed to fetch StreamerProfile from DB for channel %v: %w", username, err)
	}

	// Copy direct fields (these are already non-JSONB)
	apiProfile.ChannelID = dbProfile.ChannelID
	apiProfile.Username = dbProfile.Username
	apiProfile.Verified = dbProfile.Verified
	apiProfile.IsBanned = dbProfile.IsBanned
	apiProfile.VodEnabled = dbProfile.VodEnabled
	apiProfile.IsAffiliate = dbProfile.IsAffiliate
	apiProfile.SubscriptionEnabled = dbProfile.SubscriptionEnabled
	apiProfile.Bio = dbProfile.Bio
	apiProfile.City = dbProfile.City
	apiProfile.State = dbProfile.State
	apiProfile.TikTok = dbProfile.TikTok
	apiProfile.Country = dbProfile.Country
	apiProfile.Discord = dbProfile.Discord
	apiProfile.Twitter = dbProfile.Twitter
	apiProfile.YouTube = dbProfile.YouTube
	apiProfile.Facebook = dbProfile.Facebook
	apiProfile.Instagram = dbProfile.Instagram
	apiProfile.ProfilePic = dbProfile.ProfilePic

	var followersTimeline []models.FollowersCountPoint
	if len(dbProfile.FollowersCount) > 0 {
		if err := json.Unmarshal(dbProfile.FollowersCount, &followersTimeline); err != nil {
			log.Printf("Warning: Failed to unmarshal FollowersCount for channel %d from DB: %v", dbProfile.ChannelID, err)
			followersTimeline = []models.FollowersCountPoint{}
		}
	} else {
		followersTimeline = []models.FollowersCountPoint{}
	}
	apiProfile.FollowersCount = followersTimeline

	var livestreamUUIDs []uuid.UUID
	if len(dbProfile.Livestreams) > 0 {
		if err := json.Unmarshal(dbProfile.Livestreams, &livestreamUUIDs); err != nil {
			log.Printf("Warning: Failed to unmarshal Livestreams for channel %d from DB: %v", dbProfile.ChannelID, err)
			livestreamUUIDs = []uuid.UUID{}
		}
	} else {
		livestreamUUIDs = []uuid.UUID{}
	}

	// Fetch associated LivestreamReports and their SpamReports
	var fetchedReports []FullLivestreamReportForProfile
	if len(livestreamUUIDs) > 0 {
		var reports []models.LivestreamReport
		if err := db.DB.Where("id IN (?)", livestreamUUIDs).Order("report_start_time DESC").Find(&reports).Error; err != nil {
			log.Printf("Warning: Failed to fetch LivestreamReports for channel %d: %v", dbProfile.ChannelID, err)
		} else {
			fetchedReports = make([]FullLivestreamReportForProfile, 0, len(reports))
			for _, report := range reports {
				fullReport := FullLivestreamReportForProfile{
					LivestreamReportRestructured: LivestreamReportRestructured{
						LivestreamID:          int(report.LivestreamID),
						ReportStartTime:       report.ReportStartTime,
						DurationMinutes:       report.DurationMinutes,
						AverageViewers:        report.AverageViewers,
						PeakViewers:           report.PeakViewers,
						LowestViewers:         report.LowestViewers,
						Engagement:            report.Engagement,
						TotalMessages:         report.TotalMessages,
						UniqueChatters:        report.UniqueChatters,
						MessagesFromApps:      report.MessagesFromApps,
						ViewerCountsTimeline:  report.ViewerCountsTimeline,
						MessageCountsTimeline: report.MessageCountsTimeline,
						CreatedAt:             report.CreatedAt,
					},
				}
				if report.SpamReportID != nil {
					var spamReport models.SpamReport
					if err := db.DB.Where("id = ?", report.SpamReportID).First(&spamReport).Error; err != nil {
						log.Printf("Warning: Failed to fetch spam report %s for report %s: %v", report.SpamReportID.String(), report.ID.String(), err)

					} else {

						fullReport.SpamReport = SpamReportRestructured{
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
				fetchedReports = append(fetchedReports, fullReport)
			}
		}
	}
	apiProfile.Livestreams = fetchedReports

	return apiProfile, nil

}
