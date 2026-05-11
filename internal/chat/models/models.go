package chat_models

import (
	"time"
	"github.com/google/uuid"
)

type MessageStatus string
type MessageType string

const (
	Sent      MessageStatus = "sent"
	Delivered MessageStatus = "delivered"
	Read      MessageStatus = "read"

	TextMessage  MessageType = "text"
	VoiceMessage MessageType = "voice"
	ImageMessage MessageType = "image"
	VideoMessage MessageType = "video"
	AudioMessage MessageType = "audio"
	FileMessage  MessageType = "file"
)

type ChatMember struct {
	ChatId     uuid.UUID `json:"chat_id"`
	UserId     uuid.UUID `json:"user_id"`
	JoinedTime time.Time `json:"joined_time"`
}

type Direct struct {
	Id                uuid.UUID  `json:"id"`
	CreationTime      time.Time  `json:"creation_time"`
	FirstMember       ChatMember `json:"first_member"`
	SecondMember      ChatMember `json:"second_member"`
	CompanionName     string     `json:"companion_name"`
	CompanionNickname string     `json:"companion_nickname"`
	CompanionPhotoUrl string     `json:"companion_photo_url"`
}

type ReplyToMessage struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	FileName string `json:"file_name,omitempty"`
	IsMine   bool   `json:"is_mine"`
}

type Message struct {
	ID         uuid.UUID       `json:"id"`
	ChatID     uuid.UUID       `json:"chat_id"`
	SenderID   uuid.UUID       `json:"sender_id"`
	Content    string          `json:"content"`
	Status     MessageStatus   `json:"status"`
	CreatedAT  time.Time       `json:"created_at"`
	Type       MessageType     `json:"type"`
	Transcript string          `json:"transcript,omitempty"`
	Duration   int             `json:"duration,omitempty"`
	FileName   string          `json:"file_name,omitempty"`
	FileSize   int64           `json:"file_size,omitempty"`
	ReplyTo    *ReplyToMessage `json:"reply_to,omitempty"`
}

type DirectListItem struct {
	Id                   uuid.UUID `json:"id"`
	CreationTime         time.Time `json:"creation_time"`
	FirstUserId          uuid.UUID `json:"first_user_id"`
	SecondUserId         uuid.UUID `json:"second_user_id"`
	CompanionId          uuid.UUID `json:"companion_id"`
	CompanionName        string    `json:"companion_name"`
	CompanionNickname    string    `json:"companion_nickname"`
	CompanionPhotoURL    string    `json:"companion_photo_url"`
	CompanionAvatarColor string    `json:"companion_avatar_color"`
	LastMessageContent   string    `json:"last_message_content"`
	LastMessageAt        time.Time `json:"last_message_at"`
	LastMessageSenderId  uuid.UUID `json:"last_message_sender_id"`  
    LastMessageStatus    string    `json:"last_message_status"`
}