package chat_handler

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
	"github.com/gorilla/websocket"
	chat_service "github.com/svlynx/messenger/internal/chat/direct/service"
	chat_models "github.com/svlynx/messenger/internal/chat/models"

	"github.com/svlynx/messenger/internal/chat/ws"
	"github.com/svlynx/messenger/internal/push"
)

type DirectHandler struct {
	srvc *chat_service.DirectService
	hub  *ws.Hub
}

func NewDirectHandler(srvc *chat_service.DirectService, hub *ws.Hub) *DirectHandler {
	return &DirectHandler{
		srvc: srvc,
		hub:  hub,
	}
}

type MessageHandler struct {
	srvc       *chat_service.MessageService
	hub        *ws.Hub
	pushSender PushSender
}

type PushSender interface {
	SendToUser(ctx context.Context, userID string, payload push.PushPayload) error
}

func NewMessageHandler(srvc *chat_service.MessageService, hub *ws.Hub, pushSender PushSender) *MessageHandler {
	return &MessageHandler{
		srvc:       srvc,
		hub:        hub,
		pushSender: pushSender,
	}
}

func (h *DirectHandler) CreateNewDirectHandler(ctx *gin.Context) {
	var input chat_service.CreatedDirect
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.srvc.CreateNewDirectService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.hub != nil {
		rawPayload, _ := json.Marshal(map[string]any{
			"chat_id":      chat.Id,
			"recipient_id": input.SecondUserID,
		})
		responsePayload, _ := json.Marshal(map[string]any{
			"type":    "new_chat",
			"payload": json.RawMessage(rawPayload),
		})
		h.hub.SendToUser(input.SecondUserID, responsePayload)
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "direct created",
		"direct":  chat,
	})
}

func (h *DirectHandler) GetDirectByIdHandler(ctx *gin.Context) {
	myID, err := uuid.Parse(ctx.Query("my_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid my_id"})
		return
	}

	companionID, err := uuid.Parse(ctx.Query("companion_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid companion_id"})
		return
	}

	direct, err := h.srvc.GetDirectById(ctx.Request.Context(), myID, companionID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"direct": direct})
}

func (h *DirectHandler) GetListOfDirectsByIDHandler(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	directs, err := h.srvc.GetListOfDirectsByIDService(ctx.Request.Context(), userID)
	if err != nil {
		slog.Error("get directs error", "error", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"directs": directs})
}

func (h *MessageHandler) SendMessageHandler(ctx *gin.Context) {
	var input struct {
		ChatID      uuid.UUID               `json:"chat_id" binding:"required"`
		SenderID    uuid.UUID               `json:"sender_id" binding:"required"`
		RecipientID uuid.UUID               `json:"recipient_id" binding:"required"`
		Content     string                  `json:"content" binding:"required"`
		Type        chat_models.MessageType `json:"type"`
		SenderName  string    `json:"sender_name"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.srvc.SendMessage(ctx.Request.Context(), chat_service.CreatedMessage{
		ChatID:   input.ChatID,
		SenderID: input.SenderID,
		Content:  input.Content,
		Type:     input.Type,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.hub != nil {
		payload, err := json.Marshal(map[string]any{
			"type": "send_message",
			"payload": map[string]any{
				"id":         message.ID,
				"chat_id":    message.ChatID,
				"sender_id":  message.SenderID,
				"content":    message.Content,
				"created_at": message.CreatedAT,
				"type":       message.Type,
			},
		})
		if err == nil {
			h.hub.SendToUser(input.RecipientID, payload)
		}

		if !h.hub.IsOnline(input.RecipientID) && h.pushSender != nil {
			title := "Новое сообщение"
			if input.SenderName != "" {
				title = input.SenderName
			}
			slog.Info("push title", "senderName", input.SenderName, "title", title)
			go h.pushSender.SendToUser(context.Background(), input.RecipientID.String(), push.PushPayload{
				Title: title,
				Body:  input.Content,
				Icon:  "/favicon.png",
			})
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": message,
	})
}

func (h *MessageHandler) GetMessagesByChatIdHandler(ctx *gin.Context) {
	chatID, err := uuid.Parse(ctx.Query("chat_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	before := time.Now()
	if raw := ctx.Query("before"); raw != "" {
		before, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			before, err = time.Parse(time.RFC3339, raw)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid before"})
				return
			}
		}
	}

	limit := 50
	if raw := ctx.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
	}

	messages, err := h.srvc.GetMessagesByChatIdService(ctx.Request.Context(), chatID, before, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) SearchMessageHandler(ctx *gin.Context) {
	chatID, err := uuid.Parse(ctx.Query("chat_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	content := strings.TrimSpace(ctx.Query("content"))
	if content == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content cannot be empty"})
		return
	}

	messages, err := h.srvc.SearchMesaageService(ctx.Request.Context(), chatID, content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) UpdateMessageStatusHandler(ctx *gin.Context) {
	messageID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}

	var input struct {
		Status chat_models.MessageStatus `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.UpdateMessageStatusService(ctx.Request.Context(), input.Status, messageID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *MessageHandler) DeleteMessageHandler(ctx *gin.Context) {
	messageID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}

	if err := h.srvc.DeleteMessageService(ctx.Request.Context(), messageID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}

func (h *DirectHandler) SearchUsersHandler(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	currentUserID := strings.TrimSpace(ctx.Query("user_id"))

	users, err := h.srvc.SearchUsersService(ctx.Request.Context(), currentUserID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *MessageHandler) MarkChatMessagesAsReadHandler(ctx *gin.Context) {
	chatID, err := uuid.Parse(ctx.Query("chat_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chatid"})
		return
	}

	userIDRaw := ctx.Query("user_id")
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userid"})
		return
	}

	if err := h.srvc.MarkChatMessagesAsReadService(ctx.Request.Context(), chatID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "messages marked as read"})
}

func (h *DirectHandler) DeleteDirectHandler(ctx *gin.Context) {
	chatID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	recipientIDStr := ctx.Query("recipient_id")
	recipientID, _ := uuid.Parse(recipientIDStr)

	if err := h.srvc.DeleteDirectService(ctx.Request.Context(), chatID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.hub != nil && recipientID != uuid.Nil {
		rawPayload, _ := json.Marshal(map[string]any{
			"chat_id": chatID,
		})
		responsePayload, _ := json.Marshal(map[string]any{
			"type":    "delete_chat",
			"payload": json.RawMessage(rawPayload),
		})
		h.hub.SendToUser(recipientID, responsePayload)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "chat deleted"})
}

func (h *DirectHandler) GetUserStatusHandler(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	online, lastSeen, err := h.srvc.GetUserStatusService(ctx.Request.Context(), userID)
	slog.Info("get status", "user", userID, "online", online, "lastSeen", lastSeen, "err", err)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"online": false, "last_seen": nil})
		return
	}

	var lastSeenResp interface{}
	if !lastSeen.IsZero() {
		lastSeenResp = lastSeen
	}

	ctx.JSON(http.StatusOK, gin.H{
		"online":    online,
		"last_seen": lastSeenResp,
	})
}

type WsHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewWsHandler(hub *ws.Hub) *WsHandler {
	return &WsHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WsHandler) ServeWs(ctx *gin.Context) {
	rawUserID := ctx.Query("userid")
	if rawUserID == "" {
		rawUserID = ctx.Query("user_id")
	}

	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userid"})
		return
	}

	conn, err := h.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		slog.Error("ws upgrade error", "error", err)
		return
	}

	client := ws.NewClient(userID, conn, make(chan []byte, 256), h.hub)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *MessageHandler) SendVoiceMessageHandler(ctx *gin.Context) {
    var input struct {
        ChatID      uuid.UUID `json:"chat_id" binding:"required"`
        SenderID    uuid.UUID `json:"sender_id" binding:"required"`
        RecipientID uuid.UUID `json:"recipient_id" binding:"required"`
        AudioURL    string    `json:"audio_url" binding:"required"`
		Duration	int       `json:"duration"`
    }

    if err := ctx.ShouldBindJSON(&input); err != nil {
		slog.Error("voice send error", "error", err.Error())
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    message, err := h.srvc.SendMessage(ctx.Request.Context(), chat_service.CreatedMessage{
        ChatID:   input.ChatID,
        SenderID: input.SenderID,
        Content:  input.AudioURL,
        Type:     chat_models.VoiceMessage,
		Duration: input.Duration,
    })
    if err != nil {
		slog.Error("voice send error", "error", err.Error())
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if h.hub != nil {
        payload, _ := json.Marshal(map[string]any{
            "type": "send_message",
            "payload": map[string]any{
                "id":         message.ID,
                "chat_id":    message.ChatID,
                "sender_id":  message.SenderID,
                "content":    message.Content,
                "type":       message.Type,
                "created_at": message.CreatedAT,
            },
        })
        h.hub.SendToUser(input.RecipientID, payload)
    }

    ctx.JSON(http.StatusCreated, gin.H{"message": message})
}
