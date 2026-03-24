package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const pubSubChannel = "ws:notifications"

// Message targets a notification payload at a specific user.
type Message struct {
	UserID  uuid.UUID
	Payload []byte
}

// Hub maintains all active WebSocket clients and fans out messages.
type Hub struct {
	mu         sync.RWMutex
	clients    map[uuid.UUID][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	rdb        *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID][]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan Message, 256),
		rdb:        rdb,
	}
}

// BroadcastToUser publishes a message to Redis pub/sub so all hub instances deliver it.
func (h *Hub) BroadcastToUser(ctx context.Context, userID uuid.UUID, payload []byte) {
	type pubSubMsg struct {
		UserID  string `json:"user_id"`
		Payload []byte `json:"payload"`
	}
	data, _ := json.Marshal(pubSubMsg{UserID: userID.String(), Payload: payload})
	if err := h.rdb.Publish(ctx, pubSubChannel, data).Err(); err != nil {
		log.Printf("ws hub: publish to redis: %v", err)
	}
}

// Run processes hub events and Redis pub/sub messages. Call as a goroutine.
func (h *Hub) Run(ctx context.Context) {
	sub := h.rdb.Subscribe(ctx, pubSubChannel)
	defer sub.Close()
	subCh := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			list := h.clients[client.UserID]
			for i, c := range list {
				if c == client {
					h.clients[client.UserID] = append(list[:i], list[i+1:]...)
					close(client.send)
					break
				}
			}
			if len(h.clients[client.UserID]) == 0 {
				delete(h.clients, client.UserID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.deliverToUser(msg.UserID, msg.Payload)

		case redisMsg, ok := <-subCh:
			if !ok {
				return
			}
			type pubSubMsg struct {
				UserID  string `json:"user_id"`
				Payload []byte `json:"payload"`
			}
			var m pubSubMsg
			if err := json.Unmarshal([]byte(redisMsg.Payload), &m); err != nil {
				log.Printf("ws hub: unmarshal pubsub msg: %v", err)
				continue
			}
			uid, err := uuid.Parse(m.UserID)
			if err != nil {
				continue
			}
			h.deliverToUser(uid, m.Payload)
		}
	}
}

func (h *Hub) deliverToUser(userID uuid.UUID, payload []byte) {
	h.mu.RLock()
	list := make([]*Client, len(h.clients[userID]))
	copy(list, h.clients[userID])
	h.mu.RUnlock()

	for _, c := range list {
		select {
		case c.send <- payload:
		default:
			// client send buffer full — drop rather than block
		}
	}
}
