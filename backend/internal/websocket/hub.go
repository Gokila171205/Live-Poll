package websocket

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to specific poll rooms.
type Hub struct {
	// Registered clients partitioned by poll ID: pollID -> map[*Client]bool
	pollRooms map[string]map[*Client]bool

	// Inbound register requests
	register chan *subscription

	// Inbound unregister requests
	unregister chan *subscription

	// Inbound broadcast messages targeted to a poll room
	broadcast chan *pollBroadcast

	mu sync.RWMutex
}

type subscription struct {
	pollID string
	client *Client
}

type pollBroadcast struct {
	pollID  string
	payload []byte
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		pollRooms:  make(map[string]map[*Client]bool),
		register:   make(chan *subscription),
		unregister: make(chan *subscription),
		broadcast:  make(chan *pollBroadcast, 256),
	}
}

// Run executes the hub's main coordination loop.
func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.register:
			h.mu.Lock()
			room, exists := h.pollRooms[sub.pollID]
			if !exists {
				room = make(map[*Client]bool)
				h.pollRooms[sub.pollID] = room
			}
			room[sub.client] = true
			h.mu.Unlock()
			log.Printf("[DEBUG] Client registered to poll %s (room size: %d)", sub.pollID, len(room))

		case sub := <-h.unregister:
			h.mu.Lock()
			if room, exists := h.pollRooms[sub.pollID]; exists {
				if _, ok := room[sub.client]; ok {
					delete(room, sub.client)
					sub.client.CloseSend()
					if len(room) == 0 {
						delete(h.pollRooms, sub.pollID)
					}
					log.Printf("[DEBUG] Client unregistered from poll %s (remaining: %d)", sub.pollID, len(room))
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room, exists := h.pollRooms[msg.pollID]
			var slowClients []*Client
			if exists {
				for client := range room {
					select {
					case client.send <- msg.payload:
					default:
						// Client buffer is saturated; stage for cleanup
						slowClients = append(slowClients, client)
					}
				}
			}
			h.mu.RUnlock()

			if len(slowClients) > 0 {
				h.mu.Lock()
				if r, ok := h.pollRooms[msg.pollID]; ok {
					for _, sc := range slowClients {
						delete(r, sc)
						sc.CloseSend()
					}
					if len(r) == 0 {
						delete(h.pollRooms, msg.pollID)
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

// RegisterClient adds a client to a poll room.
func (h *Hub) RegisterClient(pollID string, client *Client) {
	h.register <- &subscription{pollID: pollID, client: client}
}

// UnregisterClient removes a client from a poll room.
func (h *Hub) UnregisterClient(pollID string, client *Client) {
	h.unregister <- &subscription{pollID: pollID, client: client}
}

// BroadcastToPoll sends a message payload to all clients currently connected to the specified poll.
func (h *Hub) BroadcastToPoll(pollID string, payload []byte) {
	h.broadcast <- &pollBroadcast{pollID: pollID, payload: payload}
}

// GetClientCount returns the number of active clients in a poll room.
func (h *Hub) GetClientCount(pollID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, exists := h.pollRooms[pollID]; exists {
		return len(room)
	}
	return 0
}
