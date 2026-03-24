package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Origin check is handled by CORS middleware on the router
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a single WebSocket connection belonging to one user.
type Client struct {
	UserID uuid.UUID
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

// ServeWS upgrades the HTTP connection to WebSocket and registers the client.
// Auth: JWT passed as ?token=<access_token> query parameter.
func ServeWS(hub *Hub, cfg *core.Config, w http.ResponseWriter, r *http.Request) {
	rawToken := r.URL.Query().Get("token")
	claims, err := core.ParseAccessToken(rawToken, cfg.SecretKey)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	client := &Client{
		UserID: userID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump reads from the WebSocket. We ignore incoming messages but need the
// pump running to detect disconnects and handle pong responses.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
