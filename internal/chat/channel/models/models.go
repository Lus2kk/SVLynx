package channel_models


import (
	"time"

	"github.com/google/uuid"
)

type ChannelRole string
type ChannelType string

const (
	RoleOwner  ChannelRole = "owner"
	RoleAdmin  ChannelRole = "admin"
	RoleEditor ChannelRole = "editor"
	RoleMember ChannelRole = "member"

	
	ChannelPublic  ChannelType = "public"
	ChannelPrivate ChannelType = "private"
)


type Channel struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Handle      string      `json:"handle"` // уникальный @handle
	Description string      `json:"description"`
	AvatarURL   string      `json:"avatar_url"`
	AvatarColor string      `json:"avatar_color"`
	Type        ChannelType `json:"type"`
	OwnerID     uuid.UUID   `json:"owner_id"`
	CreatedAt   time.Time   `json:"created_at"`
	MemberCount int         `json:"member_count"`
}


type ChannelMember struct {
	ChannelID  uuid.UUID   `json:"channel_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Role       ChannelRole `json:"role"`
	JoinedAt   time.Time   `json:"joined_at"`
	InvitedBy  *uuid.UUID  `json:"invited_by,omitempty"`
	CustomName string      `json:"custom_name,omitempty"`
}


type ChannelPost struct {
	ID        uuid.UUID  `json:"id"`
	ChannelID uuid.UUID  `json:"channel_id"`
	AuthorID  uuid.UUID  `json:"author_id"`
	Content   string     `json:"content"`
	MediaURL  string     `json:"media_url,omitempty"`
	MediaType string     `json:"media_type,omitempty"` 
	FileName  string     `json:"file_name,omitempty"`
	FileSize  int64      `json:"file_size,omitempty"`
	Pinned    bool       `json:"pinned"`
	ViewCount int        `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
}


type ChannelListItem struct {
	ID              uuid.UUID   `json:"id"`
	Name            string      `json:"name"`
	Handle          string      `json:"handle"`
	Description     string      `json:"description"`
	AvatarURL       string      `json:"avatar_url"`
	AvatarColor     string      `json:"avatar_color"`
	Type            ChannelType `json:"type"`
	MemberCount     int         `json:"member_count"`
	UserRole        ChannelRole `json:"user_role"`
	LastPostContent string      `json:"last_post_content"`
	LastPostAt      *time.Time  `json:"last_post_at"`
}


type ChannelInviteLink struct {
	ID        uuid.UUID  `json:"id"`
	ChannelID uuid.UUID  `json:"channel_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
	UseCount  int        `json:"use_count"`
	CreatedAt time.Time  `json:"created_at"`
	Active    bool       `json:"active"`
}