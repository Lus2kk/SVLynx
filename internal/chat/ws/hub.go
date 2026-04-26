package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_service"
)

type Hub struct {
	clients    map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mutex      sync.RWMutex
	Mservice   *chat_service.MessageService
	Dservice   *chat_service.DirectService
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

func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	h.mutex.RLock()
	client, ok := h.clients[userID]
	h.mutex.RUnlock()
	if !ok {
		return
	}

	select {
	case client.send <- message:
	default:
		slog.Warn("client buffer full", "user_id", userID)
	}
}

func (h *Hub) BroadcastExcept(exceptUserID uuid.UUID, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for userID, client := range h.clients {
		if userID == exceptUserID {
			continue
		}

		select {
		case client.send <- message:
		default:
			slog.Warn("client buffer full", "user_id", userID)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.id] = client
			h.mutex.Unlock()

			if err := h.Dservice.SetUserOnline(context.Background(), client.id); err != nil {
				slog.Error("failed to set user online", "user_id", client.id, "error", err)
			}

			onlinePayload, err := json.Marshal(map[string]any{
				"type": "user_online",
				"payload": map[string]any{
					"user_id": client.id.String(),
				},
			})
			if err == nil {
				h.BroadcastExcept(client.id, onlinePayload)
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			_, existed := h.clients[client.id]
			if existed {
				delete(h.clients, client.id)
				close(client.send)
			}
			h.mutex.Unlock()

			if !existed {
				continue
			}

			if err := h.Dservice.SetUserOffline(context.Background(), client.id); err != nil {
				slog.Error("failed to set user offline", "user_id", client.id, "error", err)
			}

			lastSeen := time.Now().UTC()

			offlinePayload, err := json.Marshal(map[string]any{
				"type": "user_offline",
				"payload": map[string]any{
					"user_id":   client.id.String(),
					"last_seen": lastSeen.Format(time.RFC3339),
				},
			})
			if err == nil {
				h.BroadcastExcept(client.id, offlinePayload)
			}

		case message := <-h.broadcast:
			var event BaseMessagePayload
			if err := json.Unmarshal(message, &event); err != nil {
				slog.Error("error unmarshal payload", "error", err)
				continue
			}

			switch event.Type {
			case SendMessage:
				var payload SendMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal send_message payload", "error", err)
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
					slog.Error("error sending message", "error", err)
					continue
				}

				rawPayload, err := json.Marshal(NewMessagePayload{
					ID:        message.ID,
					ChatID:    message.ChatID,
					SenderID:  message.SenderID,
					Content:   message.Content,
					CreatedAT: message.CreatedAT,
				})
				if err != nil {
					slog.Error("error marshaling new message payload", "error", err)
					continue
				}

				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    SendMessage,
					Payload: rawPayload,
				})
				if err != nil {
					slog.Error("error marshaling base message payload", "error", err)
					continue
				}

				h.SendToUser(payload.RecipientID, responsePayload)
			}
		}
	}
}