package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

const statsPushInterval = 2 * time.Second

// GET /api/events/:eventId/stats/ws?token=<jwt>
//
// Upgrades to WebSocket and pushes a StatsResponse payload every 2 seconds
// until the client disconnects or the ticker is stopped.
//
// Authentication uses the "token" query parameter because browsers cannot
// send custom headers when opening a WebSocket connection.
func (h *StatsHandler) StreamStats(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "missing token"})
		return
	}

	userID, err := service.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "invalid or expired token"})
		return
	}

	eventID := c.Param("eventId")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ws] upgrade failed for event %s: %v", eventID, err)
		return
	}
	defer conn.Close()

	// Send the first frame immediately so the client doesn't wait 2 s.
	if err := h.sendStats(conn, eventID, userID); err != nil {
		return
	}

	ticker := time.NewTicker(statsPushInterval)
	defer ticker.Stop()

	// Read loop: drain any client-side pings / detect disconnect.
	closeCh := make(chan struct{})
	go func() {
		defer close(closeCh)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-closeCh:
			return
		case <-ticker.C:
			if err := h.sendStats(conn, eventID, userID); err != nil {
				return
			}
		}
	}
}

func (h *StatsHandler) sendStats(conn *websocket.Conn, eventID, userID string) error {
	stats, err := h.statsService.GetStats(context.Background(), eventID, userID)
	if err != nil {
		payload, _ := json.Marshal(gin.H{"error": err.Error()})
		_ = conn.WriteMessage(websocket.TextMessage, payload)
		return err
	}

	payload, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, payload)
}
