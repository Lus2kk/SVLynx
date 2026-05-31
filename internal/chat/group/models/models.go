package group_models

import (
	"time"

	"github.com/google/uuid"
)

type GroupRole string
type GroupType string
type TopicType string
type GroupMessageType string
type GroupMessageStatus string

const (
	RoleCreator GroupRole = "creator"
	RoleAdmin   GroupRole = "admin"
	RoleMember  GroupRole = "member"

	GroupPublic  GroupType = "public"
	GroupPrivate GroupType = "private"

	TopicOpen  TopicType = "open"
	TopicClosed TopicType = "closed"

	GroupMsgText  GroupMessageType = "text"
	GroupMsgVoice GroupMessageType = "voice"
	GroupMsgImage GroupMessageType = "image"
	GroupMsgVideo GroupMessageType = "video"
	GroupMsgAudio GroupMessageType = "audio"
	GroupMsgFile  GroupMessageType = "file"

	GroupMsgSent      GroupMessageStatus = "sent"
	GroupMsgDelivered GroupMessageStatus = "delivered"
	GroupMsgRead      GroupMessageStatus = "read"
)

type Group struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Handle      string    `json:"handle"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatar_url"`
	AvatarColor string    `json:"avatar_color"`
	Type        GroupType `json:"type"`
	CreatorID   uuid.UUID `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	MemberCount int       `json:"member_count"`
}

type GroupMember struct {
	GroupID   uuid.UUID `json:"group_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      GroupRole  `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	CustomName string   `json:"custom_name,omitempty"`
	IsBanned  bool      `json:"is_banned"`
	BannedUntil *time.Time `json:"banned_until,omitempty"`
}

type GroupTopic struct {
	ID          uuid.UUID `json:"id"`
	GroupID     uuid.UUID `json:"group_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsClosed    bool      `json:"is_closed"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type GroupMessage struct {
	ID         uuid.UUID               `json:"id"`
	GroupID    uuid.UUID               `json:"group_id"`
	TopicID    *uuid.UUID              `json:"topic_id,omitempty"`
	SenderID   uuid.UUID               `json:"sender_id"`
	Content    string                  `json:"content"`
	Type       GroupMessageType        `json:"type"`
	Status     GroupMessageStatus      `json:"status"`
	MediaURL   string                  `json:"media_url,omitempty"`
	MediaType  string                  `json:"media_type,omitempty"`
	FileName   string                  `json:"file_name,omitempty"`
	FileSize   int64                   `json:"file_size,omitempty"`
	Duration   int                     `json:"duration,omitempty"`
	Transcript string                  `json:"transcript,omitempty"`
	ReplyToID  *uuid.UUID              `json:"reply_to_id,omitempty"`
	ReplyTo    *GroupReplyToMessage    `json:"reply_to,omitempty"`
	Pinned     bool                    `json:"pinned"`
	CreatedAt  time.Time               `json:"created_at"`
	EditedAt   *time.Time              `json:"edited_at,omitempty"`
}

type GroupReplyToMessage struct {
	ID      uuid.UUID        `json:"id"`
	Content string           `json:"content"`
	Type    GroupMessageType `json:"type"`
	IsMine  bool             `json:"is_mine"`
}

type GroupListItem struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Handle          string    `json:"handle"`
	Description     string    `json:"description"`
	AvatarURL       string    `json:"avatar_url"`
	AvatarColor     string    `json:"avatar_color"`
	Type            GroupType `json:"type"`
	MemberCount     int       `json:"member_count"`
	UserRole        GroupRole `json:"user_role"`
	LastMessageContent string  `json:"last_message_content"`
	LastMessageAt    *time.Time `json:"last_message_at"`
}

type GroupInviteLink struct {
	ID        uuid.UUID  `json:"id"`
	GroupID   uuid.UUID  `json:"group_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
	UseCount  int        `json:"use_count"`
	CreatedAt time.Time  `json:"created_at"`
	Active    bool       `json:"active"`
}

type GroupBan struct {
	GroupID    uuid.UUID  `json:"group_id"`
	UserID     uuid.UUID  `json:"user_id"`
	BannedBy   uuid.UUID  `json:"banned_by"`
	Reason     string     `json:"reason"`
	BannedAt   time.Time  `json:"banned_at"`
	Until      *time.Time `json:"until,omitempty"`
}