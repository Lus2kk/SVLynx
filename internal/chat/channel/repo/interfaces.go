package channel_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	channel_models "github.com/svlynx/messenger/internal/chat/channel/models"
)

type ChannelRepo interface {

	CreateChannelRepo(ctx context.Context, channel *channel_models.Channel) (*channel_models.Channel, error)
	GetChannelByIDRepo(ctx context.Context, channelID uuid.UUID) (*channel_models.Channel, error)
	GetChannelByHandleRepo(ctx context.Context, handle string) (*channel_models.Channel, error)
	UpdateChannelRepo(ctx context.Context, channel *channel_models.Channel) (*channel_models.Channel, error)
	DeleteChannelRepo(ctx context.Context, channelID uuid.UUID) error
	SearchChannelsRepo(ctx context.Context, query string, limit int) ([]*channel_models.Channel, error)

	
	AddMemberRepo(ctx context.Context, member *channel_models.ChannelMember) error
	RemoveMemberRepo(ctx context.Context, channelID, userID uuid.UUID) error
	GetMemberRepo(ctx context.Context, channelID, userID uuid.UUID) (*channel_models.ChannelMember, error)
	GetMembersRepo(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*channel_models.ChannelMember, error)
	UpdateMemberRoleRepo(ctx context.Context, channelID, userID uuid.UUID, role channel_models.ChannelRole) error
	GetUserChannelsRepo(ctx context.Context, userID uuid.UUID) ([]*channel_models.ChannelListItem, error)
	GetMemberCountRepo(ctx context.Context, channelID uuid.UUID) (int, error)

	
	CreatePostRepo(ctx context.Context, post *channel_models.ChannelPost) (*channel_models.ChannelPost, error)
	GetPostByIDRepo(ctx context.Context, postID uuid.UUID) (*channel_models.ChannelPost, error)
	GetPostsByChannelRepo(ctx context.Context, channelID uuid.UUID, before time.Time, limit int) ([]*channel_models.ChannelPost, error)
	UpdatePostRepo(ctx context.Context, post *channel_models.ChannelPost) (*channel_models.ChannelPost, error)
	DeletePostRepo(ctx context.Context, postID uuid.UUID) error
	PinPostRepo(ctx context.Context, postID uuid.UUID, pinned bool) error
	GetPinnedPostsRepo(ctx context.Context, channelID uuid.UUID) ([]*channel_models.ChannelPost, error)
	SearchPostsRepo(ctx context.Context, channelID uuid.UUID, query string) ([]*channel_models.ChannelPost, error)
	IncrementViewCountRepo(ctx context.Context, postID uuid.UUID) error

	
	CreateInviteLinkRepo(ctx context.Context, link *channel_models.ChannelInviteLink) (*channel_models.ChannelInviteLink, error)
	GetInviteLinkByTokenRepo(ctx context.Context, token string) (*channel_models.ChannelInviteLink, error)
	GetInviteLinksByChannelRepo(ctx context.Context, channelID uuid.UUID) ([]*channel_models.ChannelInviteLink, error)
	IncrementInviteUsageRepo(ctx context.Context, linkID uuid.UUID) error
	DeactivateInviteLinkRepo(ctx context.Context, linkID uuid.UUID) error
}