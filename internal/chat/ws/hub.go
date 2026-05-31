package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	channel_repo "github.com/svlynx/messenger/internal/chat/channel/repo"
	group_repo "github.com/svlynx/messenger/internal/chat/group/repo"
	chat_service "github.com/svlynx/messenger/internal/chat/direct/service"
)

type Hub struct {
	clients    map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mutex      sync.RWMutex
	Mservice   *chat_service.MessageService
	Dservice   *chat_service.DirectService
	CRepo      channel_repo.ChannelRepo
	GRepo      group_repo.GroupRepo
}

func NewHub(Mservice *chat_service.MessageService, Dservice *chat_service.DirectService, CRepo channel_repo.ChannelRepo, GRepo group_repo.GroupRepo) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		Mservice:   Mservice,
		Dservice:   Dservice,
		CRepo:      CRepo,
		GRepo:      GRepo,
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

func (h *Hub) broadcastToChannelMembers(ctx context.Context, channelID uuid.UUID, exceptUserID uuid.UUID, message []byte) {
	const pageSize = 200
	offset := 0

	for {
		members, err := h.CRepo.GetMembersRepo(ctx, channelID, pageSize, offset)
		if err != nil {
			slog.Error("broadcastToChannelMembers: GetMembersRepo error", "channel_id", channelID, "error", err)
			return
		}
		if len(members) == 0 {
			break
		}

		for _, m := range members {
			if m.UserID == exceptUserID {
				continue
			}
			h.SendToUser(m.UserID, message)
		}

		if len(members) < pageSize {
			break
		}
		offset += pageSize
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
			slog.Info("broadcast received", "raw", string(message))
			var event BaseMessagePayload
			if err := json.Unmarshal(message, &event); err != nil {
				slog.Error("error unmarshal payload", "error", err)
				continue
			}
			slog.Info("event type", "type", event.Type)

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

			case DeleteMessage:
				slog.Info("delete case hit")
				var payload DeleteMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal delete_message payload", "error", err)
					continue
				}
				ctx := context.Background()
				if err := h.Mservice.DeleteMessageService(ctx, payload.ID); err != nil {
					slog.Warn("message already deleted or not found, still notifying recipient", "error", err)
				}
				rawPayload, err := json.Marshal(payload)
				if err != nil {
					slog.Error("error marshaling delete payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    DeleteMessage,
					Payload: rawPayload,
				})
				if err != nil {
					slog.Error("error marshaling base delete payload", "error", err)
					continue
				}
				h.SendToUser(payload.RecipientID, responsePayload)

			case MarkAsRead:
				var payload MarkAsReadPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal mark_as_read payload", "error", err)
					continue
				}
				ctx := context.Background()
				if err := h.Mservice.MarkChatMessagesAsReadService(ctx, payload.ChatID, payload.UserID); err != nil {
					slog.Error("error marking messages as read", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    MarkAsRead,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling mark_as_read response", "error", err)
					continue
				}
				h.SendToUser(payload.RecipientID, responsePayload)

			case Typing:
				var payload TypingPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal typing payload", "error", err)
					continue
				}
				slog.Info("sending typing to", "recipient_id", payload.RecipientID)
				rawPayload, _ := json.Marshal(payload)
				responsePayload, _ := json.Marshal(BaseMessagePayload{
					Type:    Typing,
					Payload: rawPayload,
				})
				h.SendToUser(payload.RecipientID, responsePayload)

			case NewChannelPost:
				var payload NewChannelPostPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal new_channel_post payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    NewChannelPost,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling new_channel_post response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, payload.AuthorID, responsePayload)

			case UpdateChannelPost:
				var payload UpdateChannelPostPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal update_channel_post payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    UpdateChannelPost,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling update_channel_post response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, uuid.Nil, responsePayload)

			case DeleteChannelPost:
				var payload DeleteChannelPostPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal delete_channel_post payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    DeleteChannelPost,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling delete_channel_post response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, uuid.Nil, responsePayload)

			case PinChannelPost:
				var payload PinChannelPostPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal pin_channel_post payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    PinChannelPost,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling pin_channel_post response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, uuid.Nil, responsePayload)

			case ChannelMemberJoin:
				// Уведомляем всех участников что кто-то вступил
				var payload ChannelMemberEventPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal channel_member_join payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    ChannelMemberJoin,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling channel_member_join response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, payload.UserID, responsePayload)

			case ChannelMemberLeft:
				var payload ChannelMemberEventPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal channel_member_left payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    ChannelMemberLeft,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling channel_member_left response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, payload.UserID, responsePayload)

			case ChannelDeleted:
				var payload ChannelDeletedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal channel_deleted payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    ChannelDeleted,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling channel_deleted response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, uuid.Nil, responsePayload)

			case ChannelUpdated:
				var payload ChannelUpdatedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal channel_updated payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    ChannelUpdated,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling channel_updated response", "error", err)
					continue
				}
				h.broadcastToChannelMembers(context.Background(), payload.ChannelID, uuid.Nil, responsePayload)


			case NewGroupMessage:
				var payload NewGroupMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal new_group_message payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    NewGroupMessage,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling new_group_message response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, payload.AuthorID, responsePayload)

			case UpdateGroupMessage:
				var payload UpdateGroupMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal update_group_message payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    UpdateGroupMessage,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling update_group_message response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case DeleteGroupMessage:
				var payload DeleteGroupMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal delete_group_message payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    DeleteGroupMessage,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling delete_group_message response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case PinGroupMessage:
				var payload PinGroupMessagePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal pin_group_message payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    PinGroupMessage,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling pin_group_message response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupMemberJoin:
				var payload GroupMemberEventPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_member_join payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupMemberJoin,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_member_join response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupMemberLeft:
				var payload GroupMemberEventPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_member_left payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupMemberLeft,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_member_left response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupMemberRole:
				var payload GroupMemberRolePayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_member_role payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupMemberRole,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_member_role response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupMemberBanned:
				var payload GroupMemberBannedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_member_banned payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupMemberBanned,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_member_banned response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupDeleted:
				var payload GroupDeletedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_deleted payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupDeleted,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_deleted response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupUpdated:
				var payload GroupUpdatedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_updated payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupUpdated,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_updated response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupOwnershipTransferred:
				var payload GroupOwnershipTransferredPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_ownership_transferred payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupOwnershipTransferred,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_ownership_transferred response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupTopicCreated:
				var payload GroupTopicPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_topic_created payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupTopicCreated,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_topic_created response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupTopicUpdated:
				var payload GroupTopicPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_topic_updated payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupTopicUpdated,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_topic_updated response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)

			case GroupTopicDeleted:
				var payload GroupTopicDeletedPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					slog.Error("error unmarshal group_topic_deleted payload", "error", err)
					continue
				}
				responsePayload, err := json.Marshal(BaseMessagePayload{
					Type:    GroupTopicDeleted,
					Payload: event.Payload,
				})
				if err != nil {
					slog.Error("error marshaling group_topic_deleted response", "error", err)
					continue
				}
				h.broadcastToGroupMembers(context.Background(), payload.GroupID, uuid.Nil, responsePayload)
			}
		}
	}
}

func (h *Hub) broadcastToGroupMembers(ctx context.Context, groupID uuid.UUID, exceptUserID uuid.UUID, message []byte) {
	const pageSize = 200
	offset := 0

	for {
		members, err := h.GRepo.GetMembersRepo(ctx, groupID, pageSize, offset)
		if err != nil {
			slog.Error("broadcastToGroupMembers: GetMembersRepo error", "group_id", groupID, "error", err)
			return
		}
		if len(members) == 0 {
			break
		}

		for _, m := range members {
			if m.IsBanned || m.UserID == exceptUserID {
				continue
			}
			h.SendToUser(m.UserID, message)
		}

		if len(members) < pageSize {
			break
		}
		offset += pageSize
	}
}

func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
