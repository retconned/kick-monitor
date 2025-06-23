package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MonitoredChannel struct {
	ChannelID  uint   `gorm:"primaryKey"`
	ChatroomID uint   `gorm:"unique;notnull"`
	Username   string `gorm:"unique;not null"`
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ChannelData struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"` // UUID primary key
	ChannelID uint      `gorm:"not null"`             // Link to MonitoredChannel.ID
	Data      []byte    `gorm:"type:jsonb"`           // Store as JSONB
	CreatedAt time.Time `gorm:"autoCreateTime"`       // GORM will handle creating the timestamp
}

type LivestreamData struct {
	ChannelID    uint `gorm:"primaryKey"` // Primary key part 1, Foreign Key
	LivestreamID uint `gorm:"primaryKey"` // Primary key part 2, from livestream data 'id'

	// Extracted Livestream Fields
	Slug                string    `gorm:"size:255"`
	StartTime           time.Time // Original start_time from livestream data
	SessionTitle        string    `gorm:"size:255"`
	ViewerCount         int
	LivestreamCreatedAt time.Time // Original created_at from livestream data
	Tags                []byte    `gorm:"type:jsonb"` // Store tags as JSONB
	IsLive              bool
	Duration            int
	LangISO             string    `gorm:"size:10"`
	CreatedAt           time.Time `gorm:"primaryKey;autoCreateTime"` // Primary key part 3, timestamp of the fetch

}

type ChatMessage struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"` // Message UUID from data payload
	ChatroomID      uint      `gorm:"not null"`             // Link to MonitoredChannel.ChatRoomID
	LivestreamID    *uint     `gorm:"column:livestream_id"` // Nullable foreign key, pointer to uint
	SenderID        int       `gorm:"not null"`             // Sender user ID
	SenderUsername  string    `gorm:"size:255;not null"`    // Sender username (slug)
	Event           string    `gorm:"size:255;not null"`    // WebSocket event type
	Message         string    `gorm:"type:text;not null"`   // Message content
	Metadata        []byte    `gorm:"type:jsonb"`           // Metadata as JSONB (nullable if not always present)
	MessageSendTime time.Time `gorm:"not null"`             // Original message send time from data
	CreatedAt       time.Time `gorm:"autoCreateTime"`       // Timestamp of when message was processed/saved Extracted Chat Message Fields
}

type LivestreamReport struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	LivestreamID    uint      `gorm:"not null"`
	ChannelID       uint      `gorm:"not null"`
	Username        string    `gorm:"size:255;not null"`
	ReportStartTime time.Time `gorm:"not null"`
	ReportEndTime   time.Time `gorm:"not null"`
	DurationMinutes int       `gorm:"not null"`

	// Viewer Analytics
	AverageViewers int     `gorm:"not null;default:0"`
	PeakViewers    int     `gorm:"not null;default:0"`
	LowestViewers  int     `gorm:"not null;default:0"`
	Engagement     float64 `gorm:"not null;default:0.0"`

	// Chat Metrics (spam/emote related moved to SpamReport)
	TotalMessages    int `gorm:"not null;default:0"`
	UniqueChatters   int `gorm:"not null;default:0"`
	MessagesFromApps int `gorm:"not null;default:0"`

	SpamReportID *uuid.UUID `gorm:"type:uuid"` // Moved before timelines

	// Timelines
	ViewerCountsTimeline  []byte `gorm:"type:jsonb"`
	MessageCountsTimeline []byte `gorm:"type:jsonb"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type SpamReport struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	LivestreamReportID uuid.UUID `gorm:"type:uuid;not null"`
	ChannelID          uint      `gorm:"not null"` // Redundant but useful for joins
	LivestreamID       uint      `gorm:"not null"` // Redundant but useful for joins

	// New: Moved from LivestreamReport
	MessagesWithEmotes         int `gorm:"not null;default:0"`
	MessagesMultipleEmotesOnly int `gorm:"not null;default:0"`

	// Spam Analysis Fields (extracted from JSONB)
	DuplicateMessagesCount int    `gorm:"not null;default:0"`
	RepetitivePhrasesCount int    `gorm:"not null;default:0"`
	ExactDuplicateBursts   []byte `gorm:"type:jsonb"`
	SimilarMessageBursts   []byte `gorm:"type:jsonb"`
	SuspiciousChatters     []byte `gorm:"type:jsonb"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type FollowersCountPoint struct {
	Time  time.Time `json:"time"`
	Count int       `json:"count"`
}

type StreamerProfile struct {
	ChannelID           uint            `gorm:"primaryKey;autoIncrement:false"` // FK to monitored_channels.id
	Username            string          `gorm:"size:255;not null"`
	Verified            bool            `gorm:"not null;default:false"`
	IsBanned            bool            `gorm:"not null;default:false"`
	VodEnabled          bool            `gorm:"not null;default:false"`
	IsAffiliate         bool            `gorm:"not null;default:false"`
	SubscriptionEnabled bool            `gorm:"not null;default:false"`
	FollowersCount      json.RawMessage `gorm:"type:jsonb"`
	Livestreams         []byte          `gorm:"type:jsonb"`

	Bio        string `gorm:"type:text"`
	City       string `gorm:"size:255"`
	State      string `gorm:"size:255"`
	TikTok     string `gorm:"size:255"`
	Country    string `gorm:"size:255"`
	Discord    string `gorm:"size:255"`
	Twitter    string `gorm:"size:255"`
	YouTube    string `gorm:"size:255"`
	Facebook   string `gorm:"size:255"`
	Instagram  string `gorm:"size:255"`
	ProfilePic string `gorm:"type:text"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
