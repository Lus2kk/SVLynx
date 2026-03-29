package chat_models

import (
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	Sent      MessageStatus = "sent"
	Delivered MessageStatus = "delivered"
	Read   MessageStatus = "read"
)

type ChatMember struct {
	ChatId     uuid.UUID `json:"chat_id"`
	UserId     uuid.UUID `json:"user_id"`
	JoinedTime time.Time `json:"joined_time"`
}

type Direct struct {
	Id           uuid.UUID  `json:"id"`
	CreationTime time.Time  `json:"creation_time"`
	FirstMember  ChatMember `json:"first_member"`
	SecondMember ChatMember `json:"second_member"`
}

type Message struct {
	ID        uuid.UUID     `json:"id"`
	ChatID    uuid.UUID     `json:"chat_id"`
	SenderID  uuid.UUID     `json:"sender_id"`
	Content   string        `json:"content"`
	Status    MessageStatus `json:"status"`
	CreatedAT time.Time     `json:"created_at"`
}
