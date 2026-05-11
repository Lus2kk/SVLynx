package channel_service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	channel_models "github.com/svlynx/messenger/internal/chat/channel/models"
	channel_repo "github.com/svlynx/messenger/internal/chat/channel/repo"
)

type ChannelService struct {
	repo channel_repo.ChannelRepo
}

func NewChannelService(repo channel_repo.ChannelRepo) *ChannelService {
	return &ChannelService{repo: repo}
}



type CreateChannelInput struct {
	Name        string                    `json:"name"`
	Handle      string                    `json:"handle"`
	Description string                    `json:"description"`
	AvatarURL   string                    `json:"avatar_url"`
	AvatarColor string                    `json:"avatar_color"`
	Type        channel_models.ChannelType `json:"type"`
	OwnerID     uuid.UUID                 `json:"owner_id"`
}

type UpdateChannelInput struct {
	Name        string                    `json:"name"`
	Handle      string                    `json:"handle"`
	Description string                    `json:"description"`
	AvatarURL   string                    `json:"avatar_url"`
	AvatarColor string                    `json:"avatar_color"`
	Type        channel_models.ChannelType `json:"type"`
}

type CreatePostInput struct {
	ChannelID uuid.UUID `json:"channel_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url"`
	MediaType string    `json:"media_type"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
}

type UpdatePostInput struct {
	PostID    uuid.UUID `json:"post_id"`
	EditorID  uuid.UUID `json:"editor_id"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url"`
	MediaType string    `json:"media_type"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
}

type CreateInviteLinkInput struct {
	ChannelID uuid.UUID  `json:"channel_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
}



func (s *ChannelService) CreateChannelService(ctx context.Context, input CreateChannelInput) (*channel_models.Channel, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("channel name is required")
	}
	if err := validateHandle(input.Handle); err != nil {
		return nil, err
	}

	
	existing, err := s.repo.GetChannelByHandleRepo(ctx, input.Handle)
	if err != nil {
		return nil, fmt.Errorf("check handle: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("handle @%s is already taken", input.Handle)
	}

	if input.Type == "" {
		input.Type = channel_models.ChannelPublic
	}

	ch := &channel_models.Channel{
		ID:          uuid.New(),
		Name:        input.Name,
		Handle:      strings.ToLower(input.Handle),
		Description: input.Description,
		AvatarURL:   input.AvatarURL,
		AvatarColor: input.AvatarColor,
		Type:        input.Type,
		OwnerID:     input.OwnerID,
		CreatedAt:   time.Now().UTC(),
	}

	result, err := s.repo.CreateChannelRepo(ctx, ch)
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}
	return result, nil
}

func (s *ChannelService) GetChannelByIDService(ctx context.Context, channelID uuid.UUID) (*channel_models.Channel, error) {
	ch, err := s.repo.GetChannelByIDRepo(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	return ch, nil
}

func (s *ChannelService) GetChannelByHandleService(ctx context.Context, handle string) (*channel_models.Channel, error) {
	ch, err := s.repo.GetChannelByHandleRepo(ctx, strings.ToLower(handle))
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	return ch, nil
}

func (s *ChannelService) UpdateChannelService(ctx context.Context, channelID, requesterID uuid.UUID, input UpdateChannelInput) (*channel_models.Channel, error) {
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleAdmin); err != nil {
		return nil, err
	}

	ch, err := s.repo.GetChannelByIDRepo(ctx, channelID)
	if err != nil || ch == nil {
		return nil, fmt.Errorf("channel not found")
	}

	if input.Handle != "" && input.Handle != ch.Handle {
		if err := validateHandle(input.Handle); err != nil {
			return nil, err
		}
		existing, err := s.repo.GetChannelByHandleRepo(ctx, input.Handle)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != channelID {
			return nil, fmt.Errorf("handle @%s is already taken", input.Handle)
		}
		ch.Handle = strings.ToLower(input.Handle)
	}

	if input.Name != "" {
		ch.Name = input.Name
	}
	ch.Description = input.Description
	ch.AvatarURL = input.AvatarURL
	ch.AvatarColor = input.AvatarColor
	if input.Type != "" {
		ch.Type = input.Type
	}

	return s.repo.UpdateChannelRepo(ctx, ch)
}

func (s *ChannelService) DeleteChannelService(ctx context.Context, channelID, requesterID uuid.UUID) error {
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleOwner); err != nil {
		return err
	}
	return s.repo.DeleteChannelRepo(ctx, channelID)
}

func (s *ChannelService) SearchChannelsService(ctx context.Context, query string) ([]*channel_models.Channel, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	return s.repo.SearchChannelsRepo(ctx, query, 30)
}

func (s *ChannelService) JoinChannelService(ctx context.Context, channelID, userID uuid.UUID) error {
	ch, err := s.repo.GetChannelByIDRepo(ctx, channelID)
	if err != nil || ch == nil {
		return fmt.Errorf("channel not found")
	}
	if ch.Type == channel_models.ChannelPrivate {
		return fmt.Errorf("channel is private, use an invite link")
	}

	member := &channel_models.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
		Role:      channel_models.RoleMember,
		JoinedAt:  time.Now().UTC(),
	}
	return s.repo.AddMemberRepo(ctx, member)
}


func (s *ChannelService) LeaveChannelService(ctx context.Context, channelID, userID uuid.UUID) error {
	m, err := s.repo.GetMemberRepo(ctx, channelID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("you are not a member of this channel")
	}
	if m.Role == channel_models.RoleOwner {
		return fmt.Errorf("owner cannot leave the channel; transfer ownership first")
	}
	return s.repo.RemoveMemberRepo(ctx, channelID, userID)
}


func (s *ChannelService) KickMemberService(ctx context.Context, channelID, requesterID, targetID uuid.UUID) error {
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleAdmin); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, channelID, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == channel_models.RoleOwner {
		return fmt.Errorf("cannot kick the owner")
	}
	
	if target.Role == channel_models.RoleAdmin {
		requester, _ := s.repo.GetMemberRepo(ctx, channelID, requesterID)
		if requester == nil || requester.Role != channel_models.RoleOwner {
			return fmt.Errorf("only owner can kick admins")
		}
	}
	return s.repo.RemoveMemberRepo(ctx, channelID, targetID)
}


func (s *ChannelService) UpdateMemberRoleService(ctx context.Context, channelID, requesterID, targetID uuid.UUID, role channel_models.ChannelRole) error {
	if role == channel_models.RoleOwner {
		return fmt.Errorf("cannot assign owner role directly; use TransferOwnership")
	}
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleOwner); err != nil {
		return err
	}
	target, err := s.repo.GetMemberRepo(ctx, channelID, targetID)
	if err != nil || target == nil {
		return fmt.Errorf("target member not found")
	}
	if target.Role == channel_models.RoleOwner {
		return fmt.Errorf("cannot change owner role")
	}
	return s.repo.UpdateMemberRoleRepo(ctx, channelID, targetID, role)
}


func (s *ChannelService) TransferOwnershipService(ctx context.Context, channelID, ownerID, newOwnerID uuid.UUID) error {
	if err := s.requireRole(ctx, channelID, ownerID, channel_models.RoleOwner); err != nil {
		return err
	}
	newOwner, err := s.repo.GetMemberRepo(ctx, channelID, newOwnerID)
	if err != nil || newOwner == nil {
		return fmt.Errorf("new owner is not a channel member")
	}
	if err := s.repo.UpdateMemberRoleRepo(ctx, channelID, ownerID, channel_models.RoleAdmin); err != nil {
		return err
	}
	if err := s.repo.UpdateMemberRoleRepo(ctx, channelID, newOwnerID, channel_models.RoleOwner); err != nil {
		return err
	}

	ch, _ := s.repo.GetChannelByIDRepo(ctx, channelID)
	if ch != nil {
		ch.OwnerID = newOwnerID
		_, _ = s.repo.UpdateChannelRepo(ctx, ch)
	}
	return nil
}

func (s *ChannelService) GetMembersService(ctx context.Context, channelID, requesterID uuid.UUID, limit, offset int) ([]*channel_models.ChannelMember, error) {
	if err := s.requireAnyMember(ctx, channelID, requesterID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMembersRepo(ctx, channelID, limit, offset)
}

func (s *ChannelService) GetUserChannelsService(ctx context.Context, userID uuid.UUID) ([]*channel_models.ChannelListItem, error) {
	return s.repo.GetUserChannelsRepo(ctx, userID)
}



func (s *ChannelService) CreatePostService(ctx context.Context, input CreatePostInput) (*channel_models.ChannelPost, error) {
	if err := s.requireRole(ctx, input.ChannelID, input.AuthorID, channel_models.RoleEditor); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Content) == "" && input.MediaURL == "" {
		return nil, fmt.Errorf("post content or media is required")
	}

	post := &channel_models.ChannelPost{
		ID:        uuid.New(),
		ChannelID: input.ChannelID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		MediaURL:  input.MediaURL,
		MediaType: input.MediaType,
		FileName:  input.FileName,
		FileSize:  input.FileSize,
		Pinned:    false,
		ViewCount: 0,
		CreatedAt: time.Now().UTC(),
	}
	return s.repo.CreatePostRepo(ctx, post)
}

func (s *ChannelService) GetPostsService(ctx context.Context, channelID, requesterID uuid.UUID, before time.Time, limit int) ([]*channel_models.ChannelPost, error) {
	ch, err := s.repo.GetChannelByIDRepo(ctx, channelID)
	if err != nil || ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	
	if ch.Type == channel_models.ChannelPrivate {
		if err := s.requireAnyMember(ctx, channelID, requesterID); err != nil {
			return nil, err
		}
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetPostsByChannelRepo(ctx, channelID, before, limit)
}

func (s *ChannelService) UpdatePostService(ctx context.Context, input UpdatePostInput) (*channel_models.ChannelPost, error) {
	post, err := s.repo.GetPostByIDRepo(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("post not found")
	}

	member, err := s.repo.GetMemberRepo(ctx, post.ChannelID, input.EditorID)
	if err != nil || member == nil {
		return nil, fmt.Errorf("access denied")
	}

	
	if post.AuthorID != input.EditorID && !isAtLeast(member.Role, channel_models.RoleAdmin) {
		return nil, fmt.Errorf("you can only edit your own posts")
	}

	post.Content = input.Content
	post.MediaURL = input.MediaURL
	post.MediaType = input.MediaType
	post.FileName = input.FileName
	post.FileSize = input.FileSize
	return s.repo.UpdatePostRepo(ctx, post)
}

func (s *ChannelService) DeletePostService(ctx context.Context, postID, requesterID uuid.UUID) error {
	post, err := s.repo.GetPostByIDRepo(ctx, postID)
	if err != nil || post == nil {
		return fmt.Errorf("post not found")
	}

	member, err := s.repo.GetMemberRepo(ctx, post.ChannelID, requesterID)
	if err != nil || member == nil {
		return fmt.Errorf("access denied")
	}

	if post.AuthorID != requesterID && !isAtLeast(member.Role, channel_models.RoleAdmin) {
		return fmt.Errorf("you can only delete your own posts")
	}
	return s.repo.DeletePostRepo(ctx, postID)
}

func (s *ChannelService) PinPostService(ctx context.Context, postID, requesterID uuid.UUID, pinned bool) error {
	post, err := s.repo.GetPostByIDRepo(ctx, postID)
	if err != nil || post == nil {
		return fmt.Errorf("post not found")
	}
	if err := s.requireRole(ctx, post.ChannelID, requesterID, channel_models.RoleAdmin); err != nil {
		return err
	}
	return s.repo.PinPostRepo(ctx, postID, pinned)
}

func (s *ChannelService) GetPinnedPostsService(ctx context.Context, channelID, requesterID uuid.UUID) ([]*channel_models.ChannelPost, error) {
	ch, _ := s.repo.GetChannelByIDRepo(ctx, channelID)
	if ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	if ch.Type == channel_models.ChannelPrivate {
		if err := s.requireAnyMember(ctx, channelID, requesterID); err != nil {
			return nil, err
		}
	}
	return s.repo.GetPinnedPostsRepo(ctx, channelID)
}

func (s *ChannelService) SearchPostsService(ctx context.Context, channelID, requesterID uuid.UUID, query string) ([]*channel_models.ChannelPost, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	ch, _ := s.repo.GetChannelByIDRepo(ctx, channelID)
	if ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	if ch.Type == channel_models.ChannelPrivate {
		if err := s.requireAnyMember(ctx, channelID, requesterID); err != nil {
			return nil, err
		}
	}
	return s.repo.SearchPostsRepo(ctx, channelID, query)
}

func (s *ChannelService) ViewPostService(ctx context.Context, postID uuid.UUID) error {
	return s.repo.IncrementViewCountRepo(ctx, postID)
}


func (s *ChannelService) CreateInviteLinkService(ctx context.Context, input CreateInviteLinkInput) (*channel_models.ChannelInviteLink, error) {
	if err := s.requireRole(ctx, input.ChannelID, input.CreatedBy, channel_models.RoleAdmin); err != nil {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	link := &channel_models.ChannelInviteLink{
		ID:        uuid.New(),
		ChannelID: input.ChannelID,
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

func (s *ChannelService) JoinByInviteService(ctx context.Context, token string, userID uuid.UUID) (*channel_models.Channel, error) {
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

	member := &channel_models.ChannelMember{
		ChannelID: link.ChannelID,
		UserID:    userID,
		Role:      channel_models.RoleMember,
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

	return s.repo.GetChannelByIDRepo(ctx, link.ChannelID)
}

func (s *ChannelService) GetInviteLinksService(ctx context.Context, channelID, requesterID uuid.UUID) ([]*channel_models.ChannelInviteLink, error) {
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleAdmin); err != nil {
		return nil, err
	}
	return s.repo.GetInviteLinksByChannelRepo(ctx, channelID)
}

func (s *ChannelService) DeactivateInviteLinkService(ctx context.Context, linkID, channelID, requesterID uuid.UUID) error {
	if err := s.requireRole(ctx, channelID, requesterID, channel_models.RoleAdmin); err != nil {
		return err
	}
	return s.repo.DeactivateInviteLinkRepo(ctx, linkID)
}



var roleWeight = map[channel_models.ChannelRole]int{
	channel_models.RoleMember: 1,
	channel_models.RoleEditor: 2,
	channel_models.RoleAdmin:  3,
	channel_models.RoleOwner:  4,
}

func isAtLeast(have, need channel_models.ChannelRole) bool {
	return roleWeight[have] >= roleWeight[need]
}

func (s *ChannelService) requireRole(ctx context.Context, channelID, userID uuid.UUID, minRole channel_models.ChannelRole) error {
	m, err := s.repo.GetMemberRepo(ctx, channelID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("access denied: you are not a member of this channel")
	}
	if !isAtLeast(m.Role, minRole) {
		return fmt.Errorf("access denied: required role %s, you have %s", minRole, m.Role)
	}
	return nil
}

func (s *ChannelService) requireAnyMember(ctx context.Context, channelID, userID uuid.UUID) error {
	m, err := s.repo.GetMemberRepo(ctx, channelID, userID)
	if err != nil || m == nil {
		return fmt.Errorf("access denied: you are not a member of this channel")
	}
	return nil
}

func validateHandle(handle string) error {
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

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}