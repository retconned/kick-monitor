package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync" // Import sync for mutex
	"time"

	"github.com/gorilla/websocket"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/util"

	"github.com/google/uuid"
)

const (
	FetchInterval = 2 * time.Minute
	ProxyURL      = "http://localhost:8191/v1"                         // Proxy URL for API calls
	WebSocketURL  = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679" // Base WebSocket URL
	// Leeway for considering livestream data current
	LivestreamFreshnessLeeway = 20 * time.Second // 2 minutes + 20 seconds
)

// Structs for proxy response and Kick API data
type ProxyResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		URL       string            `json:"url"`
		Status    int               `json:"status"`
		Cookies   []interface{}     `json:"cookies"`
		UserAgent string            `json:"userAgent"`
		Headers   map[string]string `json:"headers"`
		Response  string            `json:"response"` // HTML content
	} `json:"solution"`
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Version        string `json:"version"`
}

type KickChannelResponse struct {
	ID                  int           `json:"id"`
	UserID              int           `json:"user_id"`
	Slug                string        `json:"slug"`
	IsBanned            bool          `json:"is_banned"`
	PlaybackURL         string        `json:"playback_url"`
	VodEnabled          bool          `json:"vod_enabled"`
	SubscriptionEnabled bool          `json:"subscription_enabled"`
	IsAffiliate         bool          `json:"is_affiliate"`
	FollowersCount      int           `json:"followers_count"`
	Following           bool          `json:"following"`
	Subscription        interface{}   `json:"subscription"`      // Or define a proper struct
	SubscriberBadges    []interface{} `json:"subscriber_badges"` // Or define a proper struct
	BannerImage         struct {
		URL string `json:"url"`
	} `json:"banner_image"`
	Livestream         *KickLivestream `json:"livestream"` // Pointer to handle null
	Role               interface{}     `json:"role"`       // Or define a proper struct
	Muted              bool            `json:"muted"`
	FollowerBadges     []interface{}   `json:"follower_badges"`      // Or define a proper struct
	OfflineBannerImage interface{}     `json:"offline_banner_image"` // Or define a proper struct
	Verified           bool            `json:"verified"`
	RecentCategories   []interface{}   `json:"recent_categories"` // Or define a proper struct
	CanHost            bool            `json:"can_host"`
	Chatroom           *KickChatroom   `json:"chatroom"` // Pointer to handle null
}

type KickLivestream struct {
	ID            int         `json:"id"`
	Slug          string      `json:"slug"`
	ChannelID     int         `json:"channel_id"`
	CreatedAt     string      `json:"created_at"` // Still string for unmarshalling
	SessionTitle  string      `json:"session_title"`
	IsLive        bool        `json:"is_live"`
	RiskLevelID   interface{} `json:"risk_level_id"`  // Or define a proper struct
	StartTime     string      `json:"start_time"`     // Still string for unmarshalling
	Source        interface{} `json:"source"`         // Or define a proper struct
	TwitchChannel interface{} `json:"twitch_channel"` // Or define a proper struct
	Duration      int         `json:"duration"`
	Language      string      `json:"language"`
	IsMature      bool        `json:"is_mature"`
	ViewerCount   int         `json:"viewer_count"`
	Thumbnail     struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	LangISO    string          `json:"lang_iso"`
	Tags       json.RawMessage `json:"tags"`       // Use json.RawMessage to keep raw JSON for tags
	Categories []interface{}   `json:"categories"` // Or define a proper struct
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
	Data    string `json:"data"` // Data field is a JSON string
}

// Struct for the ChatMessageEvent data payload (unmarshalled from the Data string)

// In-memory storage for the latest active livestream data per channel
var latestLivestream sync.Map // map[uint]LatestLivestreamInfo

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
			Color  string        `json:"color"`
			Badges []interface{} `json:"badges"`
		} `json:"identity"`
	} `json:"sender"`
	Metadata json.RawMessage `json:"metadata"` // Use json.RawMessage for metadata
	// Add other fields from the ChatMessageEvent data if needed
}

// StartMonitoringChannel initiates the data fetching and WebSocket routines for a channel.
func StartMonitoringChannel(channel *models.MonitoredChannel) {
	log.Printf("Starting monitoring for channel: %s (ID: %d)", channel.Username, channel.ID)
	latestLivestream.Store(channel.ID, LatestLivestreamInfo{}) // Start with a zero value
	// Start data fetching Go routine (uses proxy)
	go fetchDataAndPersist(channel)

	// Start WebSocket monitoring Go routine (does NOT use proxy)
	go startWebSocketMonitor(channel)
}

// FetchChannelData fetches channel data from the Kick API via the proxy.
// Optimized: This function should only be called if the channel is NOT found in the database.
func FetchChannelData(username string) (*KickChannelResponse, error) {
	log.Printf("Fetching data for channel: %s via proxy", username)
	apiURL := fmt.Sprintf("https://kick.com/api/v2/channels/%s", username)

	// Prepare the proxy request payload
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

// processChannelData fetches, prints, and persists channel and livestream data.
// Optimized: This now fetches data directly using the username from the MonitoredChannel.
func processChannelData(channel *models.MonitoredChannel) {
	// We already have the channel's basic info (ID, Username, ChatRoomID) from the database.
	// Now we fetch the full data to persist the current state.

	log.Printf("Processing data for channel: %s (ID: %d, ChatRoomID: %d)", channel.Username, channel.ID, channel.ChatRoomID)
	apiURL := fmt.Sprintf("https://kick.com/api/v2/channels/%s", channel.Username)

	// Prepare the proxy request payload
	proxyReqPayload := ProxyRequestPayload{
		Cmd:        "request.get",
		URL:        apiURL,
		MaxTimeout: 60000, // 60 seconds
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

	// Extract JSON from HTML response within the proxy's solution.response
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

	// Print channel data to console for testing
	// channelDataJSON, _ := json.MarshalIndent(kickData, "", "  ")
	log.Printf("Fetched Channel Data for %s (ID: %d, ChatRoomID: %d):\n", channel.Username, channel.ID, channel.ChatRoomID)

	// Persist channel data (using UUID as primary key)
	channelData := models.ChannelData{
		ID:        uuid.New(),         // Generate a new UUID for the primary key
		ChannelID: channel.ID,         // Use the MonitoredChannel ID as the foreign key
		Data:      []byte(jsonString), // Store the raw JSON
		// CreatedAt is handled by GORM's autoCreateTime tag
	}

	// Use Create to insert a new record for each fetch
	if err := db.DB.Create(&channelData).Error; err != nil {
		log.Printf("Error saving channel data for %s: %v", channel.Username, err)
	} else {
		log.Printf("Saved channel data for %s (Channel ID: %d, UUID: %s)", channel.Username, channel.ID, channelData.ID.String())
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
			// Handle error or set a zero time if parsing fails
		}

		// Use json.RawMessage for tags to save as JSONB directly
		tagsData := []byte{}
		if kickData.Livestream.Tags != nil {
			tagsData = kickData.Livestream.Tags
		}

		livestreamID := uint(kickData.Livestream.ID)

		livestreamData := models.LivestreamData{
			ChannelID:    channel.ID,
			LivestreamID: livestreamID, // Use the livestream ID from the data
			// CreatedAt is the fetch timestamp, handled by GORM's autoCreateTime tag

			// Populate extracted fields
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
		// Changed: Removed the OnConflict clause. Just use Create.
		if err := db.DB.Create(&livestreamData).Error; err != nil {
			log.Printf("Error saving livestream data for %s (Livestream ID: %d): %v", channel.Username, livestreamData.LivestreamID, err)
		} else {
			log.Printf("Saved livestream data for %s (Channel ID: %d, Livestream ID: %d)", channel.Username, channel.ID, livestreamData.LivestreamID)

			// Update in-memory latest livestream info
			latestLivestream.Store(channel.ID, LatestLivestreamInfo{
				LivestreamID: livestreamID,
				FetchTime:    time.Now(), // Use the current time when data was successfully fetched
				IsLive:       kickData.Livestream.IsLive,
			})
			log.Printf("Updated in-memory latest livestream for channel %s (ID: %d) to LivestreamID: %d", channel.Username, channel.ID, livestreamID)
		}
	} else {
		log.Printf("No active livestream data for channel: %s (ID: %d). Clearing in-memory latest livestream info.", channel.Username, channel.ID)
		// If no live stream data is returned, clear the in-memory info
		latestLivestream.Store(channel.ID, LatestLivestreamInfo{})
	}
}

// fetchRawProxyResponse fetches the raw proxy response.
func fetchRawProxyResponse(username string) (*ProxyResponse, error) {
	log.Printf("Fetching raw proxy response for channel: %s", username)
	apiURL := fmt.Sprintf("https://kick.com/api/v2/channels/%s", username)

	proxyReqPayload := ProxyRequestPayload{
		Cmd:        "request.get",
		URL:        apiURL,
		MaxTimeout: 60000,
	}

	proxyReqBody, err := json.Marshal(proxyReqPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling proxy request payload for raw response: %w", err)
	}

	resp, err := http.Post(ProxyURL, "application/json", bytes.NewBuffer(proxyReqBody))
	if err != nil {
		return nil, fmt.Errorf("error sending request to proxy for raw response for %s: %w", username, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading raw proxy response body for %s: %w", username, err)
	}

	var proxyResp ProxyResponse
	if err := json.Unmarshal(body, &proxyResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling raw proxy response for %s: %w", username, err)
	}

	return &proxyResp, nil
}

// createWebSocket establishes and subscribes to the WebSocket connection.
// createWebSocket establishes and subscribes to the WebSocket connection.
func createWebSocket(chatroomId uint) (*websocket.Conn, error) {
	params := url.Values{}
	params.Add("protocol", "7")
	params.Add("client", "js")
	params.Add("version", "7.4.0") // Using the version from your working code
	params.Add("flash", "false")

	fullURL := WebSocketURL + "?" + params.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}

	subscribe := map[string]interface{}{
		"event": "pusher:subscribe",
		"data": map[string]string{
			"auth":    "", // No auth required for this public channel
			"channel": fmt.Sprintf("chatrooms.%d.v2", chatroomId),
		},
	}

	if err := conn.WriteJSON(subscribe); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to subscribe to channel chatrooms.%d.v2: %w", chatroomId, err)
	}

	return conn, nil
}

// startWebSocketMonitor connects to the WebSocket, subscribes, and processes messages.
// This function does NOT use the proxy.
func startWebSocketMonitor(channel *models.MonitoredChannel) {
	for {
		conn, err := createWebSocket(channel.ChatRoomID)
		if err != nil {
			log.Printf("WebSocket connection error for channel %s (ID: %d): %v. Retrying in 5 seconds...", channel.Username, channel.ChatRoomID, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("WebSocket connected and subscribed for channel: %s (ID: %d)", channel.Username, channel.ChatRoomID)

		// Read messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error for channel %s (ID: %d): %v. Attempting to reconnect...", channel.Username, channel.ChatRoomID, err)
				conn.Close() // Close connection and break inner loop to trigger reconnect
				break
			}

			// Process and persist the message
			handleWebSocketMessage(channel, message)
		}

		// Add a small delay before attempting to reconnect after a read error
		time.Sleep(1 * time.Second)
	}
}

// handleWebSocketMessage handles incoming WebSocket messages based on their event type.
func handleWebSocketMessage(channel *models.MonitoredChannel, rawMessage []byte) {
	// Print raw WebSocket message to console for testing
	// log.Printf("Received RAW WebSocket message for %s (Channel ID: %d, ChatRoomID: %d):\n%s", channel.Username, channel.ID, channel.ChatRoomID, rawMessage)

	var msg IncomingMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error unmarshalling basic WebSocket message for %s: %v, raw message: %s", channel.Username, err, rawMessage)
		return
	}

	// Determine the relevant livestream_id
	var currentLivestreamID *uint = nil // Default to null
	if info, ok := latestLivestream.Load(channel.ID); ok {
		livestreamInfo := info.(LatestLivestreamInfo)
		// Check if the latest livestream data is recent and indicates a live stream
		if livestreamInfo.IsLive && time.Since(livestreamInfo.FetchTime) <= FetchInterval+LivestreamFreshnessLeeway {
			currentLivestreamID = &livestreamInfo.LivestreamID // Assign the livestream ID
		}
	}

	// Handle specific event types if needed
	switch msg.Event {
	case "pusher_internal:subscription_succeeded":
		log.Printf("✅ WebSocket subscription succeeded for channel: %s (ID: %d, ChatRoomID: %d)", channel.Username, channel.ID, channel.ChatRoomID)

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
			// Handle error or set a zero time if parsing fails
			// Depending on how crucial the timestamp is, you might decide to skip saving the message or log a warning.
			// For now, we'll continue but the MessageSendTime field will be the zero value of time.Time.
		}

		// Parse the message ID string into a UUID
		messageUUID, err := uuid.Parse(chatMsgData.ID)
		if err != nil {
			log.Printf("Error parsing chat message ID string into UUID for %s: %v, value: %s", channel.Username, err, chatMsgData.ID)
			// Handle error, maybe skip saving this message if ID is crucial
			return // Skip saving if UUID is invalid
		}

		// Print formatted chat message to console
		// log.Printf("💬 Chat Message from %s (ChatRoomID: %d): %s",
		// 	chatMsgData.Sender.Username,
		// 	chatMsgData.ChatroomID,
		// 	chatMsgData.Content,
		// )

		// Persist the chat message data with extracted fields
		chatMessage := models.ChatMessage{
			ID:           messageUUID,                  // Message UUID as primary key
			ChatroomID:   uint(chatMsgData.ChatroomID), // Use chatroom_id from the message data
			Event:        msg.Event,
			LivestreamID: currentLivestreamID, // Nullable livestream_id
			CreatedAt:    time.Now(),          // Timestamp of when message was processed/saved

			// Populate extracted fields
			SenderID:        chatMsgData.Sender.ID,
			SenderUsername:  chatMsgData.Sender.Slug, // Using slug for username column
			Message:         chatMsgData.Content,
			Metadata:        chatMsgData.Metadata, // Store metadata as JSONB
			MessageSendTime: messageSendTime,
		}

		if err := db.DB.Create(&chatMessage).Error; err != nil {
			log.Printf("Error saving chat message for %s (Message ID: %s): %v",
				channel.Username, chatMessage.ID.String(), err)
		} else {
			log.Printf("Saved chat message for %s (Message ID: %s, ChatRoomID: %d, LivestreamID: %v):%s",
				channel.Username, chatMessage.ID.String(), chatMessage.ChatroomID,
				func() interface{} {
					if currentLivestreamID == nil {
						return "NULL"
					}
					return *currentLivestreamID
				}(),
				chatMsgData.Content,
			)
		}

	// Add cases for other event types if you need to handle them specifically
	default:
		log.Printf("📩 Unhandled WebSocket event for %s (Channel ID: %d, ChatRoomID: %d): %s\nData: %s", channel.Username, channel.ID, channel.ChatRoomID, msg.Event, msg.Data)
	}
}
