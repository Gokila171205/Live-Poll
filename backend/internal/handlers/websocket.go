package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"

	"live-poll-backend/internal/services"
	ws "live-poll-backend/internal/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow incoming connections from any origin for polling widgets and dashboards
		return true
	},
}

// WebSocketHandler manages real-time WebSocket connections for live poll updates.
type WebSocketHandler struct {
	hub         *ws.Hub
	pollService services.PollService
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(hub *ws.Hub, pollService services.PollService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		pollService: pollService,
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket and joins the client to the poll room.
// WS /api/polls/:id/live
func (h *WebSocketHandler) ServeWS(c *gin.Context) {
	pollID := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(pollID) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	// 1. Verify that the requested poll exists before upgrading to WebSocket
	_, err := h.pollService.GetPollByID(c.Request.Context(), pollID)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify poll")
		return
	}

	// 2. Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ERROR] WebSocket upgrade failed for poll %s: %v", pollID, err)
		return
	}

	// 3. Immediately send current results snapshot so viewer does not wait for a vote
	initialResults, resErr := h.pollService.GetPollResults(c.Request.Context(), pollID)
	if resErr == nil && initialResults != nil {
		event := initialResults.ToEvent()
		if payload, jsonErr := json.Marshal(event); jsonErr == nil {
			_ = conn.WriteMessage(gorilla.TextMessage, payload)
		}
	}

	// 4. Create and register the client
	client := ws.NewClient(h.hub, pollID, conn)
	h.hub.RegisterClient(pollID, client)

	// 5. Start write pump in background; read pump blocks on current goroutine until disconnect
	go client.WritePump()
	client.ReadPump()
}
