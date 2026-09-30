package notification

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients  map[uuid.UUID]map[*websocket.Conn]bool
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

func NewHub(allowedOrigin string) *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all in development or check origin
				return true
			},
		},
	}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.clients[userID]; !exists {
		h.clients[userID] = make(map[*websocket.Conn]bool)
	}
	h.clients[userID][conn] = true
	log.Printf("🔌 WebSocket client connected: user %s (total sockets for user: %d)", userID, len(h.clients[userID]))
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if sockets, exists := h.clients[userID]; exists {
		if _, ok := sockets[conn]; ok {
			delete(sockets, conn)
			conn.Close()
			if len(sockets) == 0 {
				delete(h.clients, userID)
			}
			log.Printf("🔌 WebSocket client disconnected: user %s", userID)
		}
	}
}

// BroadcastToUser sends payload JSON to all active sockets of a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, data interface{}) {
	h.mutex.RLock()
	sockets, exists := h.clients[userID]
	if !exists || len(sockets) == 0 {
		h.mutex.RUnlock()
		return
	}

	// Copy socket references under read lock
	conns := make([]*websocket.Conn, 0, len(sockets))
	for conn := range sockets {
		conns = append(conns, conn)
	}
	h.mutex.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteJSON(data); err != nil {
			log.Printf("⚠️ Failed to write to WebSocket for user %s: %v", userID, err)
			h.Unregister(userID, conn)
		}
	}
}

// BroadcastToUsers sends payload to multiple users
func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, data interface{}) {
	for _, uid := range userIDs {
		h.BroadcastToUser(uid, data)
	}
}
