package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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