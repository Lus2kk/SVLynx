package group_service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	group_models "github.com/svlynx/messenger/internal/chat/group/models"
	group_repo "github.com/svlynx/messenger/internal/chat/group/repo"
)

type GroupService struct {
	repo group_repo.GroupRepo
}

func NewGroupService(repo group_repo.GroupRepo) *GroupService {
	return &GroupService{repo: repo}
}

type CreateGroupInput struct {
	Name        string                 `json:"name"`
	Handle      string                 `json:"handle"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	AvatarColor string                 `json:"avatar_color"`
	Type        group_models.GroupType `json:"type"`
	CreatorID   uuid.UUID              `json:"creator_id"`
}

type UpdateGroupInput struct {
	Name        string                 `json:"name"`
	Handle      string                 `json:"handle"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	AvatarColor string                 `json:"avatar_color"`
	Type        group_models.GroupType `json:"type"`
}

type CreateTopicInput struct {
	GroupID     uuid.UUID `json:"group_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsClosed    bool      `json:"is_closed"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

type UpdateTopicInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsClosed    bool   `json:"is_closed"`
}

type CreateGroupMessageInput struct {
	GroupID    uuid.UUID                      `json:"group_id"`
	TopicID    *uuid.UUID                     `json:"topic_id,omitempty"`
	SenderID   uuid.UUID                      `json:"sender_id"`
	Content    string                         `json:"content"`
	Type       group_models.GroupMessageType  `json:"type"`
	MediaURL   string                         `json:"media_url"`
	MediaType  string                         `json:"media_type"`
	FileName   string                         `json:"file_name"`
	FileSize   int64                          `json:"file_size"`
	Duration   int                            `json:"duration"`
	Transcript string                         `json:"transcript"`
	ReplyToID  *uuid.UUID                     `json:"reply_to_id,omitempty"`
}

type UpdateGroupMessageInput struct {
	MessageID  uuid.UUID `json:"message_id"`
	EditorID   uuid.UUID `json:"editor_id"`
	Content    string    `json:"content"`
	MediaURL   string    `json:"media_url"`
	MediaType  string    `json:"media_type"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
}

type CreateInviteLinkInput struct {
	GroupID   uuid.UUID  `json:"group_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
}

type BanMemberInput struct {
	GroupID  uuid.UUID  `json:"group_id"`
	UserID   uuid.UUID  `json:"user_id"`
	BannedBy uuid.UUID  `json:"banned_by"`
	Reason   string     `json:"reason"`
	Until    *time.Time `json:"until,omitempty"`
}

var groupRoleWeight = map[group_models.GroupRole]int{
	group_models.RoleMember:  1,
	group_models.RoleAdmin:   2,
	group_models.RoleCreator: 3,
}

func isAtLeastGroup(have, need group_models.GroupRole) bool {
	return groupRoleWeight[have] >= groupRoleWeight[need]
}

func (s *GroupService) requireRole(ctx context.Context, groupID, userID uuid.UUID, minRole group_models.GroupRole) error {
	m, err := s.repo.GetMemberRepo(ctx, groupID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("access denied: you are not a member of this group")
	}
	if m.IsBanned {
		return fmt.Errorf("access denied: you are banned in this group")
	}
	if !isAtLeastGroup(m.Role, minRole) {
		return fmt.Errorf("access denied: required role %s, you have %s", minRole, m.Role)
	}
	return nil
}

func (s *GroupService) requireAnyMember(ctx context.Context, groupID, userID uuid.UUID) error {
	m, err := s.repo.GetMemberRepo(ctx, groupID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("access denied: you are not a member of this group")
	}
	if m.IsBanned {
		return fmt.Errorf("access denied: you are banned in this group")
	}
	return nil
}

func validateGroupHandle(handle string) error {
	if len(handle) < 3 || len(handle) > 32 {
		return fmt.Errorf("handle must be 3–32 characters")
	}
	for _, r := range handle {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("handle may only contain letters, digits and underscores")
		}
	}
	return nil
}

func (s *GroupService) CreateGroupService(ctx context.Context, input CreateGroupInput) (*group_models.Group, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if err := validateGroupHandle(input.Handle); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetGroupByHandleRepo(ctx, input.Handle)
	if err != nil {
		return nil, fmt.Errorf("check handle: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("handle @%s is already taken", input.Handle)
	}

	if input.Type == "" {
		input.Type = group_models.GroupPublic
	}

	g := &group_models.Group{
		ID:          uuid.New(),
		Name:        input.Name,
		Handle:      strings.ToLower(input.Handle),
		Description: input.Description,
		AvatarURL:   input.AvatarURL,
		AvatarColor: input.AvatarColor,
		Type:        input.Type,
		CreatorID:   input.CreatorID,
		CreatedAt:   time.Now().UTC(),
	}

	result, err := s.repo.CreateGroupRepo(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return result, nil
}

func (s *GroupService) GetGroupByIDService(ctx context.Context, groupID uuid.UUID) (*group_models.Group, error) {
	g, err := s.repo.GetGroupByIDRepo(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	return g, nil
}

func (s *GroupService) GetGroupByHandleService(ctx context.Context, handle string) (*group_models.Group, error) {
	g, err := s.repo.GetGroupByHandleRepo(ctx, strings.ToLower(handle))
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	return g, nil
}

func (s *GroupService) UpdateGroupService(ctx context.Context, groupID, requesterID uuid.UUID, input UpdateGroupInput) (*group_models.Group, error) {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return nil, err
	}

	g, err := s.repo.GetGroupByIDRepo(ctx, groupID)
	if err != nil || g == nil {
		return nil, fmt.Errorf("group not found")
	}

	if input.Handle != "" && input.Handle != g.Handle {
		if err := validateGroupHandle(input.Handle); err != nil {
			return nil, err
		}
		existing, err := s.repo.GetGroupByHandleRepo(ctx, input.Handle)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != groupID {
			return nil, fmt.Errorf("handle @%s is already taken", input.Handle)
		}
		g.Handle = strings.ToLower(input.Handle)
	}

	if input.Name != "" {
		g.Name = input.Name
	}
	g.Description = input.Description
	g.AvatarURL = input.AvatarURL
	g.AvatarColor = input.AvatarColor
	if input.Type != "" {
		g.Type = input.Type
	}

	return s.repo.UpdateGroupRepo(ctx, g)
}

func (s *GroupService) DeleteGroupService(ctx context.Context, groupID, requesterID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleCreator); err != nil {
		return err
	}
	return s.repo.DeleteGroupRepo(ctx, groupID)
}

func (s *GroupService) SearchGroupsService(ctx context.Context, query string) ([]*group_models.Group, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	return s.repo.SearchGroupsRepo(ctx, query, 30)
}



func (s *GroupService) JoinGroupService(ctx context.Context, groupID, userID uuid.UUID) error {
	g, err := s.repo.GetGroupByIDRepo(ctx, groupID)
	if err != nil || g == nil {
		return fmt.Errorf("group not found")
	}
	if g.Type == group_models.GroupPrivate {
		return fmt.Errorf("group is private, use an invite link")
	}

	banned, err := s.repo.IsBannedRepo(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if banned {
		return fmt.Errorf("you are banned from this group")
	}

	member := &group_models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     group_models.RoleMember,
		JoinedAt: time.Now().UTC(),
	}
	return s.repo.AddMemberRepo(ctx, member)
}

func (s *GroupService) LeaveGroupService(ctx context.Context, groupID, userID uuid.UUID) error {
	m, err := s.repo.GetMemberRepo(ctx, groupID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("you are not a member of this group")
	}
	if m.Role == group_models.RoleCreator {
		return fmt.Errorf("creator cannot leave the group; transfer ownership first")
	}
	return s.repo.RemoveMemberRepo(ctx, groupID, userID)
}

func (s *GroupService) KickMemberService(ctx context.Context, groupID, requesterID, targetID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, groupID, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == group_models.RoleCreator {
		return fmt.Errorf("cannot kick the creator")
	}
	if target.Role == group_models.RoleAdmin {
		requester, _ := s.repo.GetMemberRepo(ctx, groupID, requesterID)
		if requester == nil || requester.Role != group_models.RoleCreator {
			return fmt.Errorf("only creator can kick admins")
		}
	}
	return s.repo.RemoveMemberRepo(ctx, groupID, targetID)
}

func (s *GroupService) PromoteToAdminService(ctx context.Context, groupID, requesterID, targetID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleCreator); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, groupID, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == group_models.RoleCreator {
		return fmt.Errorf("cannot change creator role")
	}
	return s.repo.UpdateMemberRoleRepo(ctx, groupID, targetID, group_models.RoleAdmin)
}

func (s *GroupService) DemoteFromAdminService(ctx context.Context, groupID, requesterID, targetID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleCreator); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, groupID, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == group_models.RoleCreator {
		return fmt.Errorf("cannot demote the creator")
	}
	return s.repo.UpdateMemberRoleRepo(ctx, groupID, targetID, group_models.RoleMember)
}

func (s *GroupService) TransferOwnershipService(ctx context.Context, groupID, creatorID, newCreatorID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, creatorID, group_models.RoleCreator); err != nil {
		return err
	}
	newCreator, err := s.repo.GetMemberRepo(ctx, groupID, newCreatorID)
	if err != nil || newCreator == nil {
		return fmt.Errorf("new creator is not a group member")
	}
	if err := s.repo.UpdateMemberRoleRepo(ctx, groupID, creatorID, group_models.RoleAdmin); err != nil {
		return err
	}
	if err := s.repo.UpdateMemberRoleRepo(ctx, groupID, newCreatorID, group_models.RoleCreator); err != nil {
		return err
	}

	g, _ := s.repo.GetGroupByIDRepo(ctx, groupID)
	if g != nil {
		g.CreatorID = newCreatorID
		_, _ = s.repo.UpdateGroupRepo(ctx, g)
	}
	return nil
}

func (s *GroupService) GetMembersService(ctx context.Context, groupID, requesterID uuid.UUID, limit, offset int) ([]*group_models.GroupMember, error) {
	if err := s.requireAnyMember(ctx, groupID, requesterID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMembersRepo(ctx, groupID, limit, offset)
}

func (s *GroupService) GetUserGroupsService(ctx context.Context, userID uuid.UUID) ([]*group_models.GroupListItem, error) {
	return s.repo.GetUserGroupsRepo(ctx, userID)
}


func (s *GroupService) BanMemberService(ctx context.Context, input BanMemberInput) error {
	if err := s.requireRole(ctx, input.GroupID, input.BannedBy, group_models.RoleAdmin); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, input.GroupID, input.UserID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == group_models.RoleCreator {
		return fmt.Errorf("cannot ban the creator")
	}
	if target.Role == group_models.RoleAdmin {
		requester, _ := s.repo.GetMemberRepo(ctx, input.GroupID, input.BannedBy)
		if requester == nil || requester.Role != group_models.RoleCreator {
			return fmt.Errorf("only creator can ban admins")
		}
	}

	ban := &group_models.GroupBan{
		GroupID:  input.GroupID,
		UserID:   input.UserID,
		BannedBy: input.BannedBy,
		Reason:   input.Reason,
		BannedAt: time.Now().UTC(),
		Until:    input.Until,
	}
	if err := s.repo.BanMemberRepo(ctx, ban); err != nil {
		return err
	}

	return s.repo.RemoveMemberRepo(ctx, input.GroupID, input.UserID)
}

func (s *GroupService) UnbanMemberService(ctx context.Context, groupID, requesterID, userID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return err
	}
	return s.repo.UnbanMemberRepo(ctx, groupID, userID)
}

func (s *GroupService) GetBansService(ctx context.Context, groupID, requesterID uuid.UUID, limit, offset int) ([]*group_models.GroupBan, error) {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetBansRepo(ctx, groupID, limit, offset)
}


func (s *GroupService) CreateTopicService(ctx context.Context, input CreateTopicInput) (*group_models.GroupTopic, error) {
	if err := s.requireRole(ctx, input.GroupID, input.CreatedBy, group_models.RoleAdmin); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	topic := &group_models.GroupTopic{
		ID:          uuid.New(),
		GroupID:     input.GroupID,
		Name:        input.Name,
		Description: input.Description,
		IsClosed:    input.IsClosed,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now().UTC(),
	}
	return s.repo.CreateTopicRepo(ctx, topic)
}

func (s *GroupService) GetTopicByIDService(ctx context.Context, topicID uuid.UUID) (*group_models.GroupTopic, error) {
	t, err := s.repo.GetTopicByIDRepo(ctx, topicID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("topic not found")
	}
	return t, nil
}

func (s *GroupService) GetTopicsByGroupService(ctx context.Context, groupID uuid.UUID) ([]*group_models.GroupTopic, error) {
	return s.repo.GetTopicsByGroupRepo(ctx, groupID)
}

func (s *GroupService) UpdateTopicService(ctx context.Context, topicID, requesterID uuid.UUID, input UpdateTopicInput) (*group_models.GroupTopic, error) {
	topic, err := s.repo.GetTopicByIDRepo(ctx, topicID)
	if err != nil || topic == nil {
		return nil, fmt.Errorf("topic not found")
	}

	if err := s.requireRole(ctx, topic.GroupID, requesterID, group_models.RoleAdmin); err != nil {
		return nil, err
	}

	if input.Name != "" {
		topic.Name = input.Name
	}
	topic.Description = input.Description
	topic.IsClosed = input.IsClosed

	return s.repo.UpdateTopicRepo(ctx, topic)
}

func (s *GroupService) DeleteTopicService(ctx context.Context, topicID, requesterID uuid.UUID) error {
	topic, err := s.repo.GetTopicByIDRepo(ctx, topicID)
	if err != nil || topic == nil {
		return fmt.Errorf("topic not found")
	}

	if err := s.requireRole(ctx, topic.GroupID, requesterID, group_models.RoleAdmin); err != nil {
		return err
	}

	return s.repo.DeleteTopicRepo(ctx, topicID)
}

func (s *GroupService) CanWriteInTopicService(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID, userID uuid.UUID) error {
	if err := s.requireAnyMember(ctx, groupID, userID); err != nil {
		return err
	}

	if topicID != nil {
		topic, err := s.repo.GetTopicByIDRepo(ctx, *topicID)
		if err != nil || topic == nil {
			return fmt.Errorf("topic not found")
		}
		if topic.IsClosed {
			member, err := s.repo.GetMemberRepo(ctx, groupID, userID)
			if err != nil || member == nil {
				return fmt.Errorf("access denied")
			}
			if !isAtLeastGroup(member.Role, group_models.RoleAdmin) {
				return fmt.Errorf("this topic is closed; only admins and the creator can write here")
			}
		}
	}
	return nil
}

func (s *GroupService) CreateGroupMessageService(ctx context.Context, input CreateGroupMessageInput) (*group_models.GroupMessage, error) {
	if err := s.CanWriteInTopicService(ctx, input.GroupID, input.TopicID, input.SenderID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Content) == "" && input.MediaURL == "" {
		return nil, fmt.Errorf("message content or media is required")
	}

	msgType := input.Type
	if msgType == "" {
		msgType = group_models.GroupMsgText
	}

	msg := &group_models.GroupMessage{
		ID:         uuid.New(),
		GroupID:    input.GroupID,
		TopicID:    input.TopicID,
		SenderID:   input.SenderID,
		Content:    input.Content,
		Type:       msgType,
		Status:     group_models.GroupMsgSent,
		MediaURL:   input.MediaURL,
		MediaType:  input.MediaType,
		FileName:   input.FileName,
		FileSize:   input.FileSize,
		Duration:   input.Duration,
		Transcript: input.Transcript,
		ReplyToID:  input.ReplyToID,
		Pinned:     false,
		CreatedAt:  time.Now().UTC(),
	}
	return s.repo.CreateGroupMessageRepo(ctx, msg)
}

func (s *GroupService) GetGroupMessagesService(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID, requesterID uuid.UUID, before time.Time, limit int) ([]*group_models.GroupMessage, error) {
	g, err := s.repo.GetGroupByIDRepo(ctx, groupID)
	if err != nil || g == nil {
		return nil, fmt.Errorf("group not found")
	}
	if g.Type == group_models.GroupPrivate {
		if err := s.requireAnyMember(ctx, groupID, requesterID); err != nil {
			return nil, err
		}
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetGroupMessagesRepo(ctx, groupID, topicID, before, limit)
}

func (s *GroupService) UpdateGroupMessageService(ctx context.Context, input UpdateGroupMessageInput) (*group_models.GroupMessage, error) {
	msg, err := s.repo.GetGroupMessageByIDRepo(ctx, input.MessageID)
	if err != nil || msg == nil {
		return nil, fmt.Errorf("message not found")
	}

	member, err := s.repo.GetMemberRepo(ctx, msg.GroupID, input.EditorID)
	if err != nil || member == nil {
		return nil, fmt.Errorf("access denied")
	}

	if msg.SenderID != input.EditorID && !isAtLeastGroup(member.Role, group_models.RoleAdmin) {
		return nil, fmt.Errorf("you can only edit your own messages")
	}

	msg.Content = input.Content
	msg.MediaURL = input.MediaURL
	msg.MediaType = input.MediaType
	msg.FileName = input.FileName
	msg.FileSize = input.FileSize
	return s.repo.UpdateGroupMessageRepo(ctx, msg)
}

func (s *GroupService) DeleteGroupMessageService(ctx context.Context, msgID, requesterID uuid.UUID) error {
	msg, err := s.repo.GetGroupMessageByIDRepo(ctx, msgID)
	if err != nil || msg == nil {
		return fmt.Errorf("message not found")
	}

	member, err := s.repo.GetMemberRepo(ctx, msg.GroupID, requesterID)
	if err != nil || member == nil {
		return fmt.Errorf("access denied")
	}

	if msg.SenderID != requesterID && !isAtLeastGroup(member.Role, group_models.RoleAdmin) {
		return fmt.Errorf("you can only delete your own messages")
	}
	return s.repo.DeleteGroupMessageRepo(ctx, msgID)
}

func (s *GroupService) PinGroupMessageService(ctx context.Context, msgID, requesterID uuid.UUID, pinned bool) error {
	msg, err := s.repo.GetGroupMessageByIDRepo(ctx, msgID)
	if err != nil || msg == nil {
		return fmt.Errorf("message not found")
	}
	if err := s.requireRole(ctx, msg.GroupID, requesterID, group_models.RoleAdmin); err != nil {
		return err
	}
	return s.repo.PinGroupMessageRepo(ctx, msgID, pinned)
}

func (s *GroupService) GetPinnedGroupMessagesService(ctx context.Context, groupID uuid.UUID, topicID *uuid.UUID, requesterID uuid.UUID) ([]*group_models.GroupMessage, error) {
	g, _ := s.repo.GetGroupByIDRepo(ctx, groupID)
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	if g.Type == group_models.GroupPrivate {
		if err := s.requireAnyMember(ctx, groupID, requesterID); err != nil {
			return nil, err
		}
	}
	return s.repo.GetPinnedGroupMessagesRepo(ctx, groupID, topicID)
}

func (s *GroupService) SearchGroupMessagesService(ctx context.Context, groupID, requesterID uuid.UUID, query string) ([]*group_models.GroupMessage, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	g, _ := s.repo.GetGroupByIDRepo(ctx, groupID)
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	if g.Type == group_models.GroupPrivate {
		if err := s.requireAnyMember(ctx, groupID, requesterID); err != nil {
			return nil, err
		}
	}
	return s.repo.SearchGroupMessagesRepo(ctx, groupID, query)
}


func (s *GroupService) CreateInviteLinkService(ctx context.Context, input CreateInviteLinkInput) (*group_models.GroupInviteLink, error) {
	if err := s.requireRole(ctx, input.GroupID, input.CreatedBy, group_models.RoleAdmin); err != nil {
		return nil, err
	}

	token, err := generateGroupToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	link := &group_models.GroupInviteLink{
		ID:        uuid.New(),
		GroupID:   input.GroupID,
		CreatedBy: input.CreatedBy,
		Token:     token,
		ExpiresAt: input.ExpiresAt,
		MaxUses:   input.MaxUses,
		UseCount:  0,
		CreatedAt: time.Now().UTC(),
		Active:    true,
	}
	return s.repo.CreateInviteLinkRepo(ctx, link)
}

func (s *GroupService) JoinByInviteService(ctx context.Context, token string, userID uuid.UUID) (*group_models.Group, error) {
	link, err := s.repo.GetInviteLinkByTokenRepo(ctx, token)
	if err != nil || link == nil {
		return nil, fmt.Errorf("invite link not found")
	}
	if !link.Active {
		return nil, fmt.Errorf("invite link is no longer active")
	}
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invite link has expired")
	}
	if link.MaxUses != nil && link.UseCount >= *link.MaxUses {
		return nil, fmt.Errorf("invite link usage limit reached")
	}

	banned, err := s.repo.IsBannedRepo(ctx, link.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if banned {
		return nil, fmt.Errorf("you are banned from this group")
	}

	member := &group_models.GroupMember{
		GroupID:   link.GroupID,
		UserID:    userID,
		Role:      group_models.RoleMember,
		JoinedAt:  time.Now().UTC(),
		InvitedBy: &link.CreatedBy,
	}
	if err := s.repo.AddMemberRepo(ctx, member); err != nil {
		return nil, err
	}
	_ = s.repo.IncrementInviteUsageRepo(ctx, link.ID)

	if link.MaxUses != nil && link.UseCount+1 >= *link.MaxUses {
		_ = s.repo.DeactivateInviteLinkRepo(ctx, link.ID)
	}

	return s.repo.GetGroupByIDRepo(ctx, link.GroupID)
}

func (s *GroupService) GetInviteLinksService(ctx context.Context, groupID, requesterID uuid.UUID) ([]*group_models.GroupInviteLink, error) {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return nil, err
	}
	return s.repo.GetInviteLinksByGroupRepo(ctx, groupID)
}

func (s *GroupService) DeactivateInviteLinkService(ctx context.Context, linkID, groupID, requesterID uuid.UUID) error {
	if err := s.requireRole(ctx, groupID, requesterID, group_models.RoleAdmin); err != nil {
		return err
	}
	return s.repo.DeactivateInviteLinkRepo(ctx, linkID)
}

func generateGroupToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}