package ws

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_models"
)

	type EventType string  

	const (
	SendMessage         EventType = "send_message"
	UpdateMessageStatus EventType = "update_status"
	DeleteMessage       EventType = "delete_message"
	NewMessage          EventType = "new_message"
	ErrorEvent          EventType = "error"
	)

	type BaseMessagePayload struct {
    Type    EventType           `json:"type"`
	Payload json.RawMessage     `json:"payload"`
	}

	type NewMessagePayload struct {
	ID        uuid.UUID         `json:"id"`
	ChatID    uuid.UUID         `json:"chat_id"`
	SenderID  uuid.UUID         `json:"sender_id"`
	Content   string            `json:"content"`
	CreatedAT time.Time         `json:"created_at"`
	}

	type SendMessagePayload struct {
	 ChatID   uuid.UUID         `json:"chat_id"`
	 SenderId uuid.UUID         `json:"sender_id"`
	 RecipientID uuid.UUID      `json:"recipient_id"`
	 Content  string            `json:"content"`
	}

	type DeleteMessagePayload struct {
    ID          uuid.UUID       `json:"id"`
	RecipientID uuid.UUID       `json:"recipient_id"`
    }

	type UpdateMessageStatusPayload struct {
	ID          uuid.UUID			                    `json:"id"`
	RecipientID uuid.UUID                               `json:"recipient_id"`
	Status      chat_models.MessageStatus			    `json:"status"`
	} 

	type ErrorPayload struct {
	 ErrorMessage string `json:"error_message"`
	 Code         uint    `json:"error_code"`
	}

