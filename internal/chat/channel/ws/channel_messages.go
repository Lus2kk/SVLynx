package ws

import "github.com/google/uuid"

type EventType string

const (
  NewChannelPost    EventType = "new_channel_post"
  UpdateChannelPost EventType = "update_channel_post"
  DeleteChannelPost EventType = "delete_channel_post"
  PinChannelPost    EventType = "pin_channel_post"
  ChannelMemberJoin EventType = "channel_member_join"
  ChannelMemberLeft EventType = "channel_member_left"
  ChannelMemberRole EventType = "channel_member_role"
  ChannelDeleted    EventType = "channel_deleted"
  ChannelUpdated    EventType = "channel_updated"
)




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