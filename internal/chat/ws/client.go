package ws

import (
	
	"log/slog"
	"time"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	
)


const (
	MaxMessageSize     = 65536 
	PongWaitingTimeout = 60 * time.Second
)


type Client struct {
	id uuid.UUID 
	connection *websocket.Conn
	send chan []byte 
	hub *Hub
}

func  NewClient (id uuid.UUID, connection *websocket.Conn, send chan []byte, hub *Hub) *Client {
	return &Client{
		id: id,
		connection: connection,
		send: send,
		hub: hub,
	}
}


func (client *Client) ReadPump() {
	defer func () {
		client.hub.unregister <- client
		client.connection.Close()
	} ()
	client.connection.SetReadLimit(MaxMessageSize)
	client.connection.SetReadDeadline(time.Now().Add(PongWaitingTimeout))
	client.connection.SetPongHandler(func(appData string) error {
		client.connection.SetReadDeadline(time.Now().Add(PongWaitingTimeout))
		return nil  
	})
	for {
		_, message, err := client.connection.ReadMessage()
		if err != nil {
			slog.Error("error of reading message ", "error", err)
			break
		}
		client.hub.broadcast <- message
	}
	}


	func (client *Client) WritePump() {
		ticker := time.NewTicker(54*time.Second)
		defer func () {
		 ticker.Stop()
		 client.connection.Close()
		}()
		for {
			select {
			case message, ok := <- client.send:
			if !ok {
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
				 if err := client.connection.WriteMessage(websocket.TextMessage, message); err != nil {
					 slog.Error("Error of sending message", "error", err.Error())
					 return
				 }

			case <-ticker.C:
				if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}

			}
		}

	}

