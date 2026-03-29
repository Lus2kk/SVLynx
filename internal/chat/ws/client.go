package ws

import (
	log "log/slog"
	"time"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	
	
)


const (
	MaxMessageSize = 512
	PongWaitingTimeout = 60 * time.Second
	WriteDeadline = 10 * time.Second
	PingInterval =54 * time.Second
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


func (client *Client) ReadPump () {
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
			log.Error("error of reading message ", "error", err)
			break
		}
		client.hub.broadcast <- message
	}
	}


		func (client *Client) WritePump() {
			ticker := time.NewTicker(PingInterval)
	defer func () {
		ticker.Stop()
		client.connection.Close()
	}()

	for {
		select {
		case message , ok := <- client.send:
		client.connection.SetWriteDeadline(time.Now().Add(WriteDeadline))
		if !ok {
			client.connection.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		writer, err := client.connection.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Error("error connecting new writer : ","error", err)
			return
		}
		writer.Write(message)

	    sended := len(client.send)
		for i := 0; i < sended; i++ {
			writer.Write([]byte{'\n'})
			writer.Write(<-client.send)
		}

		if err := writer.Close(); err!= nil {
			log.Error("error of flushing the message", "error",err)
			return
		}

	case <-ticker.C:
	client.connection.SetWriteDeadline(time.Now().Add(WriteDeadline))
	if err :=client.connection.WriteMessage(websocket.PingMessage,nil); err != nil {
		log.Error("error of sending ping", "error", err)
		return
	       }  

		}
	}
}
	
	

