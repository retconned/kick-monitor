package models

import (
	"github.com/google/uuid" // Import the uuid package
	"time"
)

// MonitoredChannel uses the Kick API's channel ID as the primary key
type MonitoredChannel struct {
	ID         uint   `gorm:"primaryKey"` // Use the Kick API channel ID
	ChatRoomID uint   `gorm:"unique;notnull"`
	Username   string `gorm:"unique;not null"`
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ChannelData uses the MonitoredChannel ID as the primary key
type ChannelData struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"` // UUID primary key
	ChannelID uint      `gorm:"not null"`             // Link to MonitoredChannel.ID
	Data      []byte    `gorm:"type:jsonb"`           // Store as JSONB
	CreatedAt time.Time `gorm:"autoCreateTime"`       // GORM will handle creating the timestamp
}

// LivestreamData with extracted fields
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

// ChatMessage with composite primary key, chatroom_id, username, and nullable livestream_id
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
