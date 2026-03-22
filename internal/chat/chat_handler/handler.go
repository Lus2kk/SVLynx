package chat_handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/chat/chat_service"
)

	type DirectHandler struct {
		srvc *chat_service.DirectService 
	} 


	func NewDirectHandler (srvc *chat_service.DirectService) *DirectHandler {
		return &DirectHandler{srvc: srvc}
	}

	
	func (h *DirectHandler) CreateNewDirectHandler (ctx *gin.Context)  {
	var input chat_service.CreatedDirect
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error with parcing" :  err.Error()})
		return 
	}
	chat, err := h.srvc.CreateNewDirectService(ctx.Request.Context(), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error with server": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H {
		"message": "New direct successfully created!",
		"direct": chat,
	})
	}


