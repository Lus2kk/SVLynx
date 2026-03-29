package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)


var Upgrader = websocket.Upgrader{
 ReadBufferSize : 1024,
 WriteBufferSize: 1024,
 CheckOrigin: func(r *http.Request) bool {
	return true 
 },
}

type WsHandler struct {
	hub *Hub
}

func NewWsHandler (hub *Hub) *WsHandler {
	return &WsHandler{hub: hub}
}

func (h *WsHandler) ConnectionHandler (ctx *gin.Context) {

	conn, err := Upgrader.Upgrade(ctx.Writer, ctx.Request,nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "connection not upgraded" })
		return
	}
	
	user_id, err := uuid.Parse(ctx.Query("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "not parsed user_id"})
		return
	}
   send := make(chan []byte, 256)
   client  := NewClient(user_id, conn, send, h.hub); 
   h.hub.register <- client
   go client.WritePump()
    client.ReadPump()

}