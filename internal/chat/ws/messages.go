package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	chat_models "github.com/svlynx/messenger/internal/chat/models"
)

type EventType string

const (
	// direct
	SendMessage         EventType = "send_message"
	UpdateMessageStatus EventType = "update_status"
	DeleteMessage       EventType = "delete_message"
	NewMessage          EventType = "new_message"
	ErrorEvent          EventType = "error"
	MarkAsRead          EventType = "mark_as_read"
	NewChat             EventType = "new_chat"
	DeleteChat          EventType = "delete_chat"
	Typing              EventType = "typing"

	// channel
	NewChannelPost    EventType = "new_channel_post"
	UpdateChannelPost EventType = "update_channel_post"
	DeleteChannelPost EventType = "delete_channel_post"
	PinChannelPost    EventType = "pin_channel_post"
	ChannelMemberJoin EventType = "channel_member_join"
	ChannelMemberLeft EventType = "channel_member_left"
	ChannelMemberRole EventType = "channel_member_role"
	ChannelDeleted    EventType = "channel_deleted"
	ChannelUpdated    EventType = "channel_updated"

	// group
	NewGroupMessage            EventType = "new_group_message"
	UpdateGroupMessage         EventType = "update_group_message"
	DeleteGroupMessage         EventType = "delete_group_message"
	PinGroupMessage            EventType = "pin_group_message"
	GroupMemberJoin            EventType = "group_member_join"
	GroupMemberLeft            EventType = "group_member_left"
	GroupMemberRole             EventType = "group_member_role"
	GroupMemberBanned          EventType = "group_member_banned"
	GroupDeleted               EventType = "group_deleted"
	GroupUpdated               EventType = "group_updated"
	GroupOwnershipTransferred  EventType = "group_ownership_transferred"
	GroupTopicCreated          EventType = "group_topic_created"
	GroupTopicUpdated          EventType = "group_topic_updated"
	GroupTopicDeleted          EventType = "group_topic_deleted"
)

type TypingPayload struct {
	ChatID      uuid.UUID `json:"chat_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
}

type DeleteChatPayload struct {
	ChatID      uuid.UUID `json:"chat_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
}

type NewChatPayload struct {
	ChatID      uuid.UUID `json:"chat_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
}

type MarkAsReadPayload struct {
	ChatID      uuid.UUID `json:"chat_id"`
	UserID      uuid.UUID `json:"user_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
}

type BaseMessagePayload struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type NewMessagePayload struct {
	ID        uuid.UUID `json:"id"`
	ChatID    uuid.UUID `json:"chat_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAT time.Time `json:"created_at"`
}

type SendMessagePayload struct {
	ChatID      uuid.UUID `json:"chat_id"`
	SenderId    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
}

type DeleteMessagePayload struct {
	ID          uuid.UUID `json:"id"`
	ChatID      uuid.UUID `json:"chat_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
}

type UpdateMessageStatusPayload struct {
	ID          uuid.UUID                 `json:"id"`
	RecipientID uuid.UUID                 `json:"recipient_id"`
	Status      chat_models.MessageStatus `json:"status"`
}

type UserOfflinePayload struct {
	UserID   uuid.UUID `json:"user_id"`
	LastSeen time.Time `json:"last_seen"`
}

// ----------------


type NewChannelPostPayload struct {
	PostID    uuid.UUID `json:"post_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url,omitempty"`
	MediaType string    `json:"media_type,omitempty"`
}


type UpdateChannelPostPayload struct {
	PostID    uuid.UUID `json:"post_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url,omitempty"`
}


type DeleteChannelPostPayload struct {
	PostID    uuid.UUID `json:"post_id"`
	ChannelID uuid.UUID `json:"channel_id"`
}


type PinChannelPostPayload struct {
	PostID    uuid.UUID `json:"post_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Pinned    bool      `json:"pinned"`
}

type ChannelMemberEventPayload struct {
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
}


type ChannelMemberRolePayload struct {
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
}


type ChannelDeletedPayload struct {
	ChannelID uuid.UUID `json:"channel_id"`
}


type ChannelUpdatedPayload struct {
	ChannelID   uuid.UUID `json:"channel_id"`
	Name        string    `json:"name"`
	Handle      string    `json:"handle"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatar_url"`
}

// ─── Group payloads ──────────────────────────────────────────────────────────

type NewGroupMessagePayload struct {
	MessageID uuid.UUID  `json:"message_id"`
	GroupID   uuid.UUID  `json:"group_id"`
	TopicID   *uuid.UUID `json:"topic_id,omitempty"`
	AuthorID  uuid.UUID  `json:"author_id"`
	Content   string     `json:"content"`
	MediaURL  string     `json:"media_url,omitempty"`
	MediaType string     `json:"media_type,omitempty"`
}

type UpdateGroupMessagePayload struct {
	MessageID uuid.UUID       `json:"message_id"`
	GroupID   uuid.UUID       `json:"group_id"`
	Content   string          `json:"content"`
	MediaURL  string          `json:"media_url,omitempty"`
	EditedAt  *time.Time      `json:"edited_at,omitempty"`
}

type DeleteGroupMessagePayload struct {
	MessageID uuid.UUID `json:"message_id"`
	GroupID   uuid.UUID `json:"group_id"`
}

type PinGroupMessagePayload struct {
	MessageID uuid.UUID `json:"message_id"`
	GroupID   uuid.UUID `json:"group_id"`
	Pinned    bool      `json:"pinned"`
}

type GroupMemberEventPayload struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
}

type GroupMemberRolePayload struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
	Role    string    `json:"role"`
}

type GroupMemberBannedPayload struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
}

type GroupDeletedPayload struct {
	GroupID uuid.UUID `json:"group_id"`
}

type GroupUpdatedPayload struct {
	GroupID      uuid.UUID `json:"group_id"`
	Name         string    `json:"name"`
	Handle       string    `json:"handle"`
	Description  string    `json:"description"`
	AvatarURL    string    `json:"avatar_url"`
}

type GroupOwnershipTransferredPayload struct {
	GroupID      uuid.UUID `json:"group_id"`
	OldCreatorID uuid.UUID `json:"old_creator_id"`
	NewCreatorID uuid.UUID `json:"new_creator_id"`
}

type GroupTopicPayload struct {
	TopicID  uuid.UUID `json:"topic_id"`
	GroupID  uuid.UUID `json:"group_id"`
	Name     string    `json:"name"`
	IsClosed bool      `json:"is_closed"`
}

type GroupTopicDeletedPayload struct {
	TopicID uuid.UUID `json:"topic_id"`
	GroupID uuid.UUID `json:"group_id"`
}