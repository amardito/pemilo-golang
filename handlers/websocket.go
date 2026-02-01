package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"pemilo/config"
	"pemilo/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin check in production
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	conn   *websocket.Conn
	roomID uint
	send   chan []byte
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *VoteUpdate
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// VoteUpdate represents a real-time vote update
type VoteUpdate struct {
	RoomID      uint `json:"room_id"`
	CandidateID uint `json:"candidate_id"`
	VoteCount   int  `json:"vote_count"`
}

var hub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan *VoteUpdate),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected to room %d", client.roomID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected from room %d", client.roomID)

		case update := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				if client.roomID == update.RoomID {
					select {
					case client.send <- marshalUpdate(update):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func marshalUpdate(update *VoteUpdate) []byte {
	return []byte(fmt.Sprintf(`{"room_id":%d,"candidate_id":%d,"vote_count":%d}`,
		update.RoomID, update.CandidateID, update.VoteCount))
}

// InitWebSocket starts the WebSocket hub
func InitWebSocket() {
	go hub.Run()
}

// HandleWebSocket handles WebSocket connections for real-time updates
func HandleWebSocket(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	var roomID uint
	if _, err := fmt.Sscanf(roomIDStr, "%d", &roomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		roomID: roomID,
		send:   make(chan []byte, 256),
	}

	hub.register <- client

	// Writer goroutine
	go func() {
		defer func() {
			conn.Close()
		}()
		for message := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				break
			}
		}
	}()

	// Reader goroutine (mainly for ping/pong)
	go func() {
		defer func() {
			hub.unregister <- client
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// BroadcastVoteUpdate sends vote update to all clients watching the room
func BroadcastVoteUpdate(roomID uint, candidateID uint) {
	// Get updated vote count
	var count int64
	config.DB.Model(&models.Vote{}).Where("candidate_id = ?", candidateID).Count(&count)

	hub.broadcast <- &VoteUpdate{
		RoomID:      roomID,
		CandidateID: candidateID,
		VoteCount:   int(count),
	}
}
