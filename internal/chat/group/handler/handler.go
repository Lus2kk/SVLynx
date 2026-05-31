package group_handler

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
	group_models "github.com/svlynx/messenger/internal/chat/group/models"
	group_service "github.com/svlynx/messenger/internal/chat/group/service"
	"github.com/svlynx/messenger/internal/chat/ws"
)

type GroupHandler struct {
	srvc *group_service.GroupService
	hub  *ws.Hub
}

func NewGroupHandler(srvc *group_service.GroupService, hub *ws.Hub) *GroupHandler {
	return &GroupHandler{srvc: srvc, hub: hub}
}

func (h *GroupHandler) broadcastToGroup(groupID uuid.UUID, eventType string, payloadData map[string]any) {
	if h.hub == nil {
		return
	}

	rawPayload, err := json.Marshal(payloadData)
	if err != nil {
		slog.Error("group broadcast: marshal payload", "error", err)
		return
	}
	envelope, err := json.Marshal(map[string]any{
		"type":    eventType,
		"payload": json.RawMessage(rawPayload),
	})
	if err != nil {
		slog.Error("group broadcast: marshal envelope", "error", err)
		return
	}

	members, err := h.srvc.GetMembersService(
		context.TODO(),
		groupID, uuid.Nil, 1000, 0,
	)
	if err != nil {
		return
	}

	for _, m := range members {
		if !m.IsBanned && h.hub.IsOnline(m.UserID) {
			h.hub.SendToUser(m.UserID, envelope)
		}
	}
}


func (h *GroupHandler) CreateGroupHandler(ctx *gin.Context) {
	var input group_service.CreateGroupInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.srvc.CreateGroupService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"group": group})
}

func (h *GroupHandler) GetGroupByIDHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	group, err := h.srvc.GetGroupByIDService(ctx.Request.Context(), groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) GetGroupByHandleHandler(ctx *gin.Context) {
	handle := strings.TrimSpace(ctx.Param("handle"))
	if handle == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "handle is required"})
		return
	}

	group, err := h.srvc.GetGroupByHandleService(ctx.Request.Context(), handle)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) UpdateGroupHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var input group_service.UpdateGroupInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.srvc.UpdateGroupService(ctx.Request.Context(), groupID, requesterID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_updated", map[string]any{
		"group_id":    group.ID,
		"name":        group.Name,
		"handle":      group.Handle,
		"description": group.Description,
		"avatar_url":  group.AvatarURL,
	})

	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) DeleteGroupHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	h.broadcastToGroup(groupID, "group_deleted", map[string]any{
		"group_id": groupID,
	})

	if err := h.srvc.DeleteGroupService(ctx.Request.Context(), groupID, requesterID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func (h *GroupHandler) SearchGroupsHandler(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	groups, err := h.srvc.SearchGroupsService(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *GroupHandler) GetUserGroupsHandler(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	groups, err := h.srvc.GetUserGroupsService(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *GroupHandler) JoinGroupHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.JoinGroupService(ctx.Request.Context(), groupID, body.UserID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_join", map[string]any{
		"group_id": groupID,
		"user_id":  body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "joined group"})
}

func (h *GroupHandler) LeaveGroupHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.LeaveGroupService(ctx.Request.Context(), groupID, body.UserID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_left", map[string]any{
		"group_id": groupID,
		"user_id":  body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "left group"})
}

func (h *GroupHandler) KickMemberHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
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

	if err := h.srvc.KickMemberService(ctx.Request.Context(), groupID, requesterID, targetID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_left", map[string]any{
		"group_id": groupID,
		"user_id":  targetID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "member kicked"})
}

func (h *GroupHandler) PromoteToAdminHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	targetID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var body struct {
		RequesterID uuid.UUID `json:"requester_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.PromoteToAdminService(ctx.Request.Context(), groupID, body.RequesterID, targetID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_role", map[string]any{
		"group_id": groupID,
		"user_id":  targetID,
		"role":     group_models.RoleAdmin,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "member promoted to admin"})
}

func (h *GroupHandler) DemoteFromAdminHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	targetID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var body struct {
		RequesterID uuid.UUID `json:"requester_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.DemoteFromAdminService(ctx.Request.Context(), groupID, body.RequesterID, targetID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_role", map[string]any{
		"group_id": groupID,
		"user_id":  targetID,
		"role":     group_models.RoleMember,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "admin demoted to member"})
}

func (h *GroupHandler) TransferOwnershipHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var body struct {
		CreatorID    uuid.UUID `json:"creator_id" binding:"required"`
		NewCreatorID uuid.UUID `json:"new_creator_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.TransferOwnershipService(ctx.Request.Context(), groupID, body.CreatorID, body.NewCreatorID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_ownership_transferred", map[string]any{
		"group_id":       groupID,
		"old_creator_id": body.CreatorID,
		"new_creator_id": body.NewCreatorID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "ownership transferred"})
}

func (h *GroupHandler) GetMembersHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	members, err := h.srvc.GetMembersService(ctx.Request.Context(), groupID, requesterID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *GroupHandler) BanMemberHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	targetID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var body struct {
		BannedBy uuid.UUID  `json:"banned_by" binding:"required"`
		Reason   string     `json:"reason"`
		Until    *time.Time `json:"until"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := group_service.BanMemberInput{
		GroupID:  groupID,
		UserID:   targetID,
		BannedBy: body.BannedBy,
		Reason:   body.Reason,
		Until:    body.Until,
	}

	if err := h.srvc.BanMemberService(ctx.Request.Context(), input); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_member_banned", map[string]any{
		"group_id": groupID,
		"user_id":  targetID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "member banned"})
}

func (h *GroupHandler) UnbanMemberHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
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

	if err := h.srvc.UnbanMemberService(ctx.Request.Context(), groupID, requesterID, targetID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "member unbanned"})
}

func (h *GroupHandler) GetBansHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	bans, err := h.srvc.GetBansService(ctx.Request.Context(), groupID, requesterID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"bans": bans})
}

func (h *GroupHandler) CreateTopicHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var input group_service.CreateTopicInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.GroupID = groupID

	topic, err := h.srvc.CreateTopicService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "group_topic_created", map[string]any{
		"topic_id": topic.ID,
		"group_id": topic.GroupID,
		"name":     topic.Name,
		"is_closed": topic.IsClosed,
	})

	ctx.JSON(http.StatusCreated, gin.H{"topic": topic})
}

func (h *GroupHandler) GetTopicsHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	topics, err := h.srvc.GetTopicsByGroupService(ctx.Request.Context(), groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (h *GroupHandler) GetTopicByIDHandler(ctx *gin.Context) {
	topicID, err := uuid.Parse(ctx.Param("topic_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic_id"})
		return
	}

	topic, err := h.srvc.GetTopicByIDService(ctx.Request.Context(), topicID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"topic": topic})
}

func (h *GroupHandler) UpdateTopicHandler(ctx *gin.Context) {
	topicID, err := uuid.Parse(ctx.Param("topic_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic_id"})
		return
	}

	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var input group_service.UpdateTopicInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic, err := h.srvc.UpdateTopicService(ctx.Request.Context(), topicID, requesterID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(topic.GroupID, "group_topic_updated", map[string]any{
		"topic_id":  topic.ID,
		"group_id":  topic.GroupID,
		"name":      topic.Name,
		"is_closed": topic.IsClosed,
	})

	ctx.JSON(http.StatusOK, gin.H{"topic": topic})
}

func (h *GroupHandler) DeleteTopicHandler(ctx *gin.Context) {
	topicID, err := uuid.Parse(ctx.Param("topic_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	topic, _ := h.srvc.GetTopicByIDService(ctx.Request.Context(), topicID)

	if err := h.srvc.DeleteTopicService(ctx.Request.Context(), topicID, requesterID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if topic != nil {
		h.broadcastToGroup(topic.GroupID, "group_topic_deleted", map[string]any{
			"topic_id":  topicID,
			"group_id":  topic.GroupID,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "topic deleted"})
}

func (h *GroupHandler) CreateGroupMessageHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var input group_service.CreateGroupMessageInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.GroupID = groupID

	msg, err := h.srvc.CreateGroupMessageService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "new_group_message", map[string]any{
		"message_id": msg.ID,
		"group_id":   msg.GroupID,
		"topic_id":   msg.TopicID,
		"sender_id":  msg.SenderID,
		"content":    msg.Content,
		"type":       msg.Type,
		"created_at": msg.CreatedAt,
	})

	ctx.JSON(http.StatusCreated, gin.H{"message": msg})
}

func (h *GroupHandler) GetGroupMessagesHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var topicID *uuid.UUID
	if tid := ctx.Query("topic_id"); tid != "" {
		parsed, err := uuid.Parse(tid)
		if err == nil {
			topicID = &parsed
		}
	}

	before := time.Now()
	if raw := ctx.Query("before"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			before = t
		}
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	messages, err := h.srvc.GetGroupMessagesService(ctx.Request.Context(), groupID, topicID, requesterID, before, limit)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *GroupHandler) UpdateGroupMessageHandler(ctx *gin.Context) {
	msgID, err := uuid.Parse(ctx.Param("message_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}

	var input group_service.UpdateGroupMessageInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.MessageID = msgID

	msg, err := h.srvc.UpdateGroupMessageService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(msg.GroupID, "update_group_message", map[string]any{
		"message_id": msg.ID,
		"group_id":   msg.GroupID,
		"content":    msg.Content,
		"edited_at":  msg.EditedAt,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": msg})
}

func (h *GroupHandler) DeleteGroupMessageHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	msgID, err := uuid.Parse(ctx.Param("message_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.srvc.DeleteGroupMessageService(ctx.Request.Context(), msgID, requesterID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "delete_group_message", map[string]any{
		"message_id": msgID,
		"group_id":   groupID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}

func (h *GroupHandler) PinGroupMessageHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	msgID, err := uuid.Parse(ctx.Param("message_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
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

	if err := h.srvc.PinGroupMessageService(ctx.Request.Context(), msgID, body.UserID, body.Pinned); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(groupID, "pin_group_message", map[string]any{
		"message_id": msgID,
		"group_id":   groupID,
		"pinned":     body.Pinned,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "pin status updated"})
}

func (h *GroupHandler) GetPinnedGroupMessagesHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var topicID *uuid.UUID
	if tid := ctx.Query("topic_id"); tid != "" {
		parsed, err := uuid.Parse(tid)
		if err == nil {
			topicID = &parsed
		}
	}

	messages, err := h.srvc.GetPinnedGroupMessagesService(ctx.Request.Context(), groupID, topicID, requesterID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *GroupHandler) SearchGroupMessagesHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	query := strings.TrimSpace(ctx.Query("q"))

	messages, err := h.srvc.SearchGroupMessagesService(ctx.Request.Context(), groupID, requesterID, query)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}


func (h *GroupHandler) CreateInviteLinkHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var input group_service.CreateInviteLinkInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.GroupID = groupID

	link, err := h.srvc.CreateInviteLinkService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"invite": link})
}

func (h *GroupHandler) GetInviteLinksHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}
	requesterID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	links, err := h.srvc.GetInviteLinksService(ctx.Request.Context(), groupID, requesterID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"invites": links})
}

func (h *GroupHandler) JoinByInviteHandler(ctx *gin.Context) {
	token := ctx.Param("token")

	var body struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.srvc.JoinByInviteService(ctx.Request.Context(), token, body.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.broadcastToGroup(group.ID, "group_member_join", map[string]any{
		"group_id": group.ID,
		"user_id":  body.UserID,
	})

	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

func (h *GroupHandler) DeactivateInviteLinkHandler(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
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

	if err := h.srvc.DeactivateInviteLinkService(ctx.Request.Context(), linkID, groupID, requesterID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "invite link deactivated"})
}