package channelhandler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	channel_models "github.com/svlynx/messenger/internal/chat/channel/channel_models"
	channel_service "github.com/svlynx/messenger/internal/chat/channel/channel_service"
	"github.com/svlynx/messenger/internal/chat/ws"
)

type ChannelHandler struct {
	srvc *channel_service.ChannelService
	hub  *ws.Hub
}

func NewChannelHandler(srvc *channel_service.ChannelService, hub *ws.Hub) *ChannelHandler {
	return &ChannelHandler{srvc: srvc, hub: hub}
}


func (h *ChannelHandler) CreateChannelHandler(ctx *gin.Context) {
	var input channel_service.CreateChannelInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.srvc.CreateChannelService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"channel": ch})
}

func (h *ChannelHandler) GetChannelByIDHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	ch, err := h.srvc.GetChannelByIDService(ctx.Request.Context(), channelID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"channel": ch})
}

func (h *ChannelHandler) GetChannelByHandleHandler(ctx *gin.Context) {
	handle := strings.TrimSpace(ctx.Param("handle"))
	if handle == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "handle is required"})
		return
	}

	ch, err := h.srvc.GetChannelByHandleService(ctx.Request.Context(), handle)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"channel": ch})
}

func (h *ChannelHandler) UpdateChannelHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var input channel_service.UpdateChannelInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.srvc.UpdateChannelService(ctx.Request.Context(), channelID, requesterID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// уведомляем всех подписчиков об изменении канала
	h.broadcastToChannel(channelID, "channel_updated", map[string]any{
		"channel_id":  ch.ID,
		"name":        ch.Name,
		"handle":      ch.Handle,
		"description": ch.Description,
		"avatar_url":  ch.AvatarURL,
	})

	ctx.JSON(http.StatusOK, gin.H{"channel": ch})
}

// DELETE /channels/:id
func (h *ChannelHandler) DeleteChannelHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// уведомляем ДО удаления, пока ещё можем получить список участников
	h.broadcastToChannel(channelID, "channel_deleted", map[string]any{
		"channel_id": channelID,
	})

	if err := h.srvc.DeleteChannelService(ctx.Request.Context(), channelID, requesterID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "channel deleted"})
}

// GET /channels/search?q=...
func (h *ChannelHandler) SearchChannelsHandler(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	channels, err := h.srvc.SearchChannelsService(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"channels": channels})
}

// GET /channels?user_id=...  — список каналов пользователя
func (h *ChannelHandler) GetUserChannelsHandler(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	channels, err := h.srvc.GetUserChannelsService(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"channels": channels})
}

// ─── Участники / Подписки ────────────────────────────────────────────────────

// POST /channels/:id/join
func (h *ChannelHandler) JoinChannelHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.JoinChannelService(ctx.Request.Context(), channelID, body.UserID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "channel_member_join", map[string]any{
		"channel_id": channelID,
		"user_id":    body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "joined channel"})
}

// POST /channels/:id/leave
func (h *ChannelHandler) LeaveChannelHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.LeaveChannelService(ctx.Request.Context(), channelID, body.UserID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "channel_member_left", map[string]any{
		"channel_id": channelID,
		"user_id":    body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "left channel"})
}

// DELETE /channels/:id/members/:user_id  — кик
func (h *ChannelHandler) KickMemberHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	targetID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("requester_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid requester_id"})
		return
	}

	if err := h.srvc.KickMemberService(ctx.Request.Context(), channelID, requesterID, targetID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "channel_member_left", map[string]any{
		"channel_id": channelID,
		"user_id":    targetID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "member kicked"})
}

// PATCH /channels/:id/members/:user_id/role
func (h *ChannelHandler) UpdateMemberRoleHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	targetID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var body struct {
		RequesterID uuid.UUID                    `json:"requester_id" binding:"required"`
		Role        channel_models.ChannelRole   `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.UpdateMemberRoleService(ctx.Request.Context(), channelID, body.RequesterID, targetID, body.Role); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "channel_member_role", map[string]any{
		"channel_id": channelID,
		"user_id":    targetID,
		"role":       body.Role,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// POST /channels/:id/transfer
func (h *ChannelHandler) TransferOwnershipHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var body struct {
		OwnerID    uuid.UUID `json:"owner_id" binding:"required"`
		NewOwnerID uuid.UUID `json:"new_owner_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.TransferOwnershipService(ctx.Request.Context(), channelID, body.OwnerID, body.NewOwnerID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ownership transferred"})
}

// GET /channels/:id/members
func (h *ChannelHandler) GetMembersHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	members, err := h.srvc.GetMembersService(ctx.Request.Context(), channelID, requesterID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"members": members})
}


func (h *ChannelHandler) CreatePostHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var input channel_service.CreatePostInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ChannelID = channelID

	post, err := h.srvc.CreatePostService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	h.broadcastToChannel(channelID, "new_channel_post", map[string]any{
		"post_id":    post.ID,
		"channel_id": post.ChannelID,
		"author_id":  post.AuthorID,
		"content":    post.Content,
		"media_url":  post.MediaURL,
		"media_type": post.MediaType,
		"created_at": post.CreatedAt,
	})

	ctx.JSON(http.StatusCreated, gin.H{"post": post})
}


func (h *ChannelHandler) GetPostsHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	before := time.Now()
	if raw := ctx.Query("before"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			before = t
		}
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	posts, err := h.srvc.GetPostsService(ctx.Request.Context(), channelID, requesterID, before, limit)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"posts": posts})
}


func (h *ChannelHandler) UpdatePostHandler(ctx *gin.Context) {
	postID, err := uuid.Parse(ctx.Param("post_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
		return
	}

	var input channel_service.UpdatePostInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.PostID = postID

	post, err := h.srvc.UpdatePostService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(post.ChannelID, "update_channel_post", map[string]any{
		"post_id":    post.ID,
		"channel_id": post.ChannelID,
		"content":    post.Content,
		"media_url":  post.MediaURL,
		"edited_at":  post.EditedAt,
	})

	ctx.JSON(http.StatusOK, gin.H{"post": post})
}


func (h *ChannelHandler) DeletePostHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	postID, err := uuid.Parse(ctx.Param("post_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.srvc.DeletePostService(ctx.Request.Context(), postID, requesterID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "delete_channel_post", map[string]any{
		"post_id":    postID,
		"channel_id": channelID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}

// PATCH /channels/:id/posts/:post_id/pin
func (h *ChannelHandler) PinPostHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	postID, err := uuid.Parse(ctx.Param("post_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
		return
	}

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Pinned bool      `json:"pinned"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.PinPostService(ctx.Request.Context(), postID, body.UserID, body.Pinned); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(channelID, "pin_channel_post", map[string]any{
		"post_id":    postID,
		"channel_id": channelID,
		"pinned":     body.Pinned,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "post pin status updated"})
}

// GET /channels/:id/posts/pinned
func (h *ChannelHandler) GetPinnedPostsHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	posts, err := h.srvc.GetPinnedPostsService(ctx.Request.Context(), channelID, requesterID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"posts": posts})
}


func (h *ChannelHandler) SearchPostsHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	query := strings.TrimSpace(ctx.Query("q"))

	posts, err := h.srvc.SearchPostsService(ctx.Request.Context(), channelID, requesterID, query)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"posts": posts})
}


func (h *ChannelHandler) ViewPostHandler(ctx *gin.Context) {
	postID, err := uuid.Parse(ctx.Param("post_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
		return
	}
	_ = h.srvc.ViewPostService(ctx.Request.Context(), postID)
	ctx.JSON(http.StatusOK, gin.H{"message": "view counted"})
}

func (h *ChannelHandler) CreateInviteLinkHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var input channel_service.CreateInviteLinkInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ChannelID = channelID

	link, err := h.srvc.CreateInviteLinkService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"invite": link})
}


func (h *ChannelHandler) GetInviteLinksHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	links, err := h.srvc.GetInviteLinksService(ctx.Request.Context(), channelID, requesterID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"invites": links})
}


func (h *ChannelHandler) JoinByInviteHandler(ctx *gin.Context) {
	token := ctx.Param("token")

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.srvc.JoinByInviteService(ctx.Request.Context(), token, body.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToChannel(ch.ID, "channel_member_join", map[string]any{
		"channel_id": ch.ID,
		"user_id":    body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"channel": ch})
}


func (h *ChannelHandler) DeactivateInviteLinkHandler(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}
	linkID, err := uuid.Parse(ctx.Param("link_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid link_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.srvc.DeactivateInviteLinkService(ctx.Request.Context(), linkID, channelID, requesterID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "invite link deactivated"})
}

func (h *ChannelHandler) broadcastToChannel(channelID uuid.UUID, eventType string, payloadData map[string]any) {
	if h.hub == nil {
		return
	}

	rawPayload, err := json.Marshal(payloadData)
	if err != nil {
		slog.Error("channel broadcast: marshal payload", "error", err)
		return
	}
	envelope, err := json.Marshal(map[string]any{
		"type":    eventType,
		"payload": json.RawMessage(rawPayload),
	})
	if err != nil {
		slog.Error("channel broadcast: marshal envelope", "error", err)
		return
	}


	members, err := h.srvc.GetMembersService(
		context.TODO(),
		channelID, uuid.Nil, 1000, 0,
	)
	if err != nil {
		return
	}

	for _, m := range members {
		if h.hub.IsOnline(m.UserID) {
			h.hub.SendToUser(m.UserID, envelope)
		}
	}
}