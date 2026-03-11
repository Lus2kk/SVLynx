package chat_models

import (
	"time"
	"github.com/google/uuid"
)


type ChatType string
const (
    Direct  ChatType = "direct"   
    Group   ChatType = "group"    
    Channel ChatType = "channel"  
)

type UserRole string 
const ( 
	Admin UserRole = "admin"
	Owner UserRole = "owner"
    Member UserRole = "member"
)

type MessageStatus string 
const (
	Sent MessageStatus = "sent"
	Delivered MessageStatus = "delivered"
	Written MessageType = "written"
)

type MessageType string 
const (
	Text MessageType = "text"
	Image MessageType = "image"
	Video MessageType = "video"
	Voice MessageType = "voice"
	Document MessageType = "document"
)


type ChatMember struct {
	Id         uuid.UUID `json:"id" db:"id"`
	ChatId     uuid.UUID `json:"chat_id" db:"chat_id"`
	UserId     uuid.UUID `json:"user_id" db:"user_id"`
	JoinedTime time.Time `json:"joined_time" db:"joined_time"`
	Role       UserRole  `json:"role" db:"role"`
}

type Chat struct {
	Id                   uuid.UUID    `json:"id" db:"id"`
	Type                 ChatType     `json:"type" db:"type"`
	Name                 string       `json:"name" db:"name"`
	AvatarURL            *string      `json:"avatar_url" db:"avatar_url"`
	OwnerID              uuid.UUID    `json:"owner_id" db:"owner_id"`
	CreationTime         time.Time    `json:"creation_time" db:"creation_time"`
	Members              []ChatMember `json:"members" db:"members"`
	UnWrittenMessageCount int         `json:"un_written_message_count" db:"un_written_message_count"`
}

type Message struct {
    ID          uuid.UUID     `db:"id"          json:"id"`
    ChatID      uuid.UUID     `db:"chat_id"      json:"chat_id"`
    SenderID    uuid.UUID     `db:"sender_id"    json:"sender_id"`
    Type        MessageType   `db:"type"         json:"type"`
    Content     *string       `db:"content"      json:"content,omitempty"`
    ReplyToID   *uuid.UUID    `db:"reply_to_id"  json:"reply_to_id,omitempty"`
    Status      MessageStatus `db:"status"       json:"status"`
    IsEdited    bool          `db:"is_edited"    json:"is_edited"`
    IsDeleted   bool          `db:"is_deleted"   json:"is_deleted"`
    CreatedAt   time.Time     `db:"created_at"   json:"created_at"`
    UpdatedAt   time.Time     `db:"updated_at"   json:"updated_at"`

    
    MediaFiles []MediaFile  `db:"-" json:"attachments,omitempty"`
    ReplyTo     *Message      `db:"-" json:"reply_to,omitempty"`
    ReadBy      []uuid.UUID   `db:"-" json:"read_by,omitempty"` 
}

type MediaFile struct {
	Id            uuid.UUID   `json:"id" db:"id"`
	MessageId     uuid.UUID   `json:"message_id" db:"message_id"`
	Type          MessageType `json:"type" db:"type"`
	FileURL       string      `json:"file_url" db:"file_url"`
	FileName      *string     `json:"file_name" db:"file_name"`
	CreateTime    time.Time   `json:"create_time" db:"create_time"`
	MediaDuration int16       `json:"media_duration" db:"media_duration"`
	Size          int64       `json:"size" db:"size"`
}