package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_service"
)



type Hub struct {
	clients map[uuid.UUID]*Client
	register chan *Client
	unregister chan *Client
	broadcast  chan []byte 
	mutex sync.RWMutex
	Mservice  *chat_service.MessageService 
	Dservice  *chat_service.DirectService
}

func NewHub(Mservice *chat_service.MessageService, Dservice *chat_service.DirectService) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		Mservice:   Mservice,
		Dservice:   Dservice,
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}



func (h *Hub) Run() {
    for {
        select {


        case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.id] = client
			h.mutex.Unlock()
			go h.Dservice.UpdateLastSeenService(context.Background(), client.id)
            
        case client := <-h.unregister:
    h.mutex.Lock()
    if _, ok := h.clients[client.id]; ok {
        delete(h.clients, client.id)
        close(client.send)
    }
    h.mutex.Unlock()
	go h.Dservice.SetUserOffline(context.Background(), client.id)

            
        case message := <-h.broadcast:
			var event BaseMessagePayload 
			if err := json.Unmarshal(message, &event); err != nil {
				slog.Error("Error unmarshal payload!", "error", err)
				continue 
			} 



			switch event.Type{

			case SendMessage: 
    var payload SendMessagePayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Error("Error unmarshal payload!", "error", err)
        continue
    }
    
    ctx := context.Background()
    input := chat_service.CreatedMessage{
        ChatID:   payload.ChatID,
        SenderID: payload.SenderId,
        Content:  payload.Content,
    }
    
    message, err := h.Mservice.SendMessage(ctx, input)
    if err != nil {
		slog.Error("Error of sending message ", "error", err)
        continue
    }
    
    responsePayload, err := json.Marshal(NewMessagePayload{
        ID:        message.ID,
        ChatID:    message.ChatID,
        SenderID:  message.SenderID,
        Content:   message.Content,
        CreatedAT: message.CreatedAT,
    })
    if err != nil {
		slog.Error("Error of marshaling message ", "error", err)
        continue
    }

    h.mutex.RLock()
    if recipient, ok := h.clients[payload.RecipientID]; ok {
        recipient.send <- responsePayload
    }
    h.mutex.RUnlock()
			


	case DeleteMessage:

	var payload DeleteMessagePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
	slog.Error("failed to unmarshal payload", "error", err)
		continue
	}
	ctx :=  context.Background()
	 if  err := h.Mservice.DeleteMessageService(ctx,payload.ID); err != nil {
	slog.Error("Error of deleting message ", "error", err)
		continue
		}
	responce := BaseMessagePayload {
	Type: DeleteMessage,
	Payload: event.Payload,
	}
	responcePayload, err := json.Marshal(responce)
	if err != nil {
	slog.Error("Error of marshaling message ", "error",err)
	continue
	}
	h.mutex.RLock()
	if recipient, ok :=h.clients[payload.RecipientID];  ok {
	recipient.send <- responcePayload
	 } 
	 h.mutex.RUnlock()
			


case UpdateMessageStatus:

	var payload UpdateMessageStatusPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Error("Error unmarshaling message", "error", err)
		continue
	}
	ctx := context.Background()

if err := h.Mservice.UpdateMessageStatusService(ctx,payload.Status,payload.ID); err != nil {
	slog.Error("Error updating message", "error", err)
		continue
	}
	recponce := BaseMessagePayload{
		Type: UpdateMessageStatus,
		Payload: event.Payload,
	}
	recponcePayload, err := json.Marshal(recponce)
	if err != nil {
	slog.Error("Error marshaling message", "error", err)
		continue
	}
	h.mutex.RLock()
		if recipient, ok := h.clients[payload.RecipientID]; ok {
			recipient.send <- recponcePayload
	}
	h.mutex.RUnlock()
			}
        }
    }
}