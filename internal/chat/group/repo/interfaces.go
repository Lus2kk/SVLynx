package group_repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	group_models "github.com/svlynx/messenger/internal/chat/group/models"
)

type GroupRepo interface {
	CreateGroupRepo(ctx context.Context, group *group_models.Group) (*group_models.Group, error)
	GetGroupByIDRepo(ctx context.Context, groupID uuid.UUID) (*group_models.Group, error)
	GetGroupByHandleRepo(ctx context.Context, handle string) (*group_models.Group, error)
	UpdateGroupRepo(ctx context.Context, group *group_models.Group) (*group_models.Group, error)
	DeleteGroupRepo(ctx context.Context, groupID uuid.UUID) error
	SearchGroupsRepo(ctx context.Context, query string, limit int) ([]*group_models.Group, error)

	AddMemberRepo(ctx context.Context, member *group_models.GroupMember) error
	RemoveMemberRepo(ctx context.Context, groupID, userID uuid.UUID) error
	GetMemberRepo(ctx context.Context, groupID, userID uuid.UUID) (*group_models.GroupMember, error)
	GetMembersRepo(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*group_models.GroupMember, error)
	UpdateMemberRoleRepo(ctx context.Context, groupID, userID uuid.UUID, role group_models.GroupRole) error
	GetUserGroupsRepo(ctx context.Context, userID uuid.UUID) ([]*group_models.GroupListItem, error)
	GetMemberCountRepo(ctx context.Context, groupID uuid.UUID) (int, error)

	BanMemberRepo(ctx context.Context, ban *group_models.GroupBan) error
	UnbanMemberRepo(ctx context.Context, groupID, userID uuid.UUID) error
	IsBannedRepo(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	GetBansRepo(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*group_models.GroupBan, error)

	CreateTopicRepo(ctx context.Context, topic *group_models.GroupTopic) (*group_models.GroupTopic, error)
	GetTopicByIDRepo(ctx context.Context, topicID uuid.UUID) (*group_models.GroupTopic, error)
	GetTopicsByGroupRepo(ctx context.Context, groupID uuid.UUID) ([]*group_models.GroupTopic, error)
	UpdateTopicRepo(ctx context.Context, topic *group_models.GroupTopic) (*group_models.GroupTopic, error)
	DeleteTopicRepo(ctx context.Context, topicID uuid.UUID) error

	CreateGroupMessageRepo(ctx context.Context, msg *group_models.GroupMessage) (*group_models.GroupMessage, error)
	GetGroupMessageByIDRepo(ctx context.Context, msgID uuid.UUID) (*group_models.GroupMessage, error)
	GetGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID, before time.Time, limit int) ([]*group_models.GroupMessage, error)
	UpdateGroupMessageRepo(ctx context.Context, msg *group_models.GroupMessage) (*group_models.GroupMessage, error)
	DeleteGroupMessageRepo(ctx context.Context, msgID uuid.UUID) error
	PinGroupMessageRepo(ctx context.Context, msgID uuid.UUID, pinned bool) error
	GetPinnedGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID) ([]*group_models.GroupMessage, error)
	SearchGroupMessagesRepo(ctx context.Context, groupID uuid.UUID, query string) ([]*group_models.GroupMessage, error)

	CreateInviteLinkRepo(ctx context.Context, link *group_models.GroupInviteLink) (*group_models.GroupInviteLink, error)
	GetInviteLinkByTokenRepo(ctx context.Context, token string) (*group_models.GroupInviteLink, error)
	GetInviteLinksByGroupRepo(ctx context.Context, groupID uuid.UUID) ([]*group_models.GroupInviteLink, error)
	IncrementInviteUsageRepo(ctx context.Context, linkID uuid.UUID) error
	DeactivateInviteLinkRepo(ctx context.Context, linkID uuid.UUID) error
}