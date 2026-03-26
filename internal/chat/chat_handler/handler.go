package chat_handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_models"
	"github.com/svlynx/messenger/internal/chat/chat_service"
)

type DirectHandler struct {
	srvc *chat_service.DirectService
}

func NewDirectHandler(srvc *chat_service.DirectService) *DirectHandler {
	return &DirectHandler{srvc: srvc}
}

type MessageHandler struct {
	srvc *chat_service.MessageService
}

func NewMessageHandler(srvc *chat_service.MessageService) *MessageHandler {
	return &MessageHandler{srvc: srvc}
}

func (h *DirectHandler) CreateNewDirectHandler(ctx *gin.Context) {
	var input chat_service.CreatedDirect
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error with parcing": err.Error()})
		return
	}
	chat, err := h.srvc.CreateNewDirectService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error with server": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "New direct successfully created!",
		"direct":  chat,
	})
}

func (h *DirectHandler) GetDirectByIdHandler(ctx *gin.Context) {
	myId, err := uuid.Parse(ctx.Query("my_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid my_id"})
		return
	}

	companionId, err := uuid.Parse(ctx.Query("companion_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid companion_id"})
		return
	}

	direct, err := h.srvc.GetDirectById(ctx.Request.Context(), myId, companionId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"direct": direct})
}

func (h *DirectHandler) GetListOfDirectsByIDHandler(ctx *gin.Context) {
	user_id, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	directs, err := h.srvc.GetListOfDirectsByIDService(ctx.Request.Context(), user_id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"directs": directs})

}

func (h *MessageHandler) SendMessageHandler(ctx *gin.Context) {
	var input chat_service.CreatedMessage
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error with parsing": err.Error()})
		return
	}
	message, err := h.srvc.SendMessage(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error with server": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"info":             "message was sended",
		"message content ": message})
}

func (h *MessageHandler) GetMessagesByChatIdHandler(ctx *gin.Context) {
	chatId, err := uuid.Parse(ctx.Query("chat_id"))
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

	messages, err := h.srvc.GetMessagesByChatIdService(ctx.Request.Context(), chatId, before, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error with server": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) SearchMessageHandler(ctx *gin.Context) {
	chat_id, err := uuid.Parse(ctx.Query("chat_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content := ctx.Query("content")
	if content == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error ": "content cannot be empty"})
		return
	}
	messages, err := h.srvc.SearchMesaageService(ctx.Request.Context(), chat_id, content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"troubles with server": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) UpdateMessageStatusHandler(ctx *gin.Context) {

	messageId, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}

	var input struct {
		Status chat_models.MessageStatus `json:"status"`
	}
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.srvc.UpdateMessageStatusService(ctx.Request.Context(), input.Status, messageId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "status updated"})
	}

	func (h *MessageHandler) DeleteMessageHandler(ctx *gin.Context) {
	messageId, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}
	if err := h.srvc.DeleteMessageService(ctx.Request.Context(), messageId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error with server": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "message deleted"})
	
	}
