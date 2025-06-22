package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/signaling-server/internal/config"
	"github.com/signaling-server/internal/middleware"
	"github.com/signaling-server/internal/service"
	"github.com/signaling-server/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

type WebSocketHandler struct {
	signalingService *service.SignalingService
	userService      *service.UserService
	config           *config.Config
	logger           *logger.Logger
}

func NewWebSocketHandler(
	signalingService *service.SignalingService,
	userService *service.UserService,
	config *config.Config,
	logger *logger.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		signalingService: signalingService,
		userService:      userService,
		config:           config,
		logger:           logger,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles signaling
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get session ID from middleware
	sessionID := middleware.GetSessionID(r)
	if sessionID == "" {
		h.logger.Error("No session ID found")
		http.Error(w, "No session found", http.StatusBadRequest)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create or get user
	ctx := context.Background()
	user, err := h.userService.GetOrCreateUser(ctx, sessionID)
	if err != nil {
		h.logger.Errorf("Failed to create user: %v", err)
		return
	}

	// Use the user ID from the created/retrieved user
	userID := user.ID

	// Add connection to signaling service
	_, err = h.signalingService.AddConnection(userID, conn, sessionID)
	if err != nil {
		h.logger.Errorf("Failed to add connection: %v", err)
		return
	}
	defer h.signalingService.RemoveConnection(userID)

	// Set connection timeouts
	conn.SetReadDeadline(time.Now().Add(time.Duration(h.config.Server.ReadTimeout) * time.Second))
	conn.SetWriteDeadline(time.Now().Add(time.Duration(h.config.Server.WriteTimeout) * time.Second))

	// Set up ping/pong handlers for connection health
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(time.Duration(h.config.Server.ReadTimeout) * time.Second))
		return nil
	})

	// Send STUN/TURN server configuration
	if err := h.sendSTUNConfig(conn); err != nil {
		h.logger.Errorf("Failed to send STUN config: %v", err)
	}

	// Handle messages
	h.handleConnection(ctx, userID, conn)
}

// handleConnection manages the WebSocket connection lifecycle
func (h *WebSocketHandler) handleConnection(ctx context.Context, userID string, conn *websocket.Conn) {
	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Channel for handling ping
	done := make(chan struct{})

	// Goroutine for sending pings
	go func() {
		defer close(done)
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					h.logger.Errorf("Failed to send ping: %v", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Main message handling loop
	for {
		// Read message
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Update read deadline
		conn.SetReadDeadline(time.Now().Add(time.Duration(h.config.Server.ReadTimeout) * time.Second))

		// Handle message
		if err := h.signalingService.HandleMessage(ctx, userID, message); err != nil {
			h.logger.Errorf("Failed to handle message from user %s: %v", userID, err)
			// Continue processing other messages instead of breaking
		}

		// Update user activity
		if err := h.userService.UpdateUserActivity(ctx, userID); err != nil {
			h.logger.Errorf("Failed to update user activity: %v", err)
		}
	}
}

// sendSTUNConfig sends STUN/TURN server configuration to the client
func (h *WebSocketHandler) sendSTUNConfig(conn *websocket.Conn) error {
	config := map[string]interface{}{
		"type": "stun_config",
		"data": map[string]interface{}{
			"iceServers": []map[string]interface{}{
				{
					"urls": []string{"stun:stun.l.google.com:19302"},
				},
				{
					"urls": []string{"stun:stun1.l.google.com:19302"},
				},
			},
		},
		"timestamp": time.Now().Unix(),
	}

	if err := conn.WriteJSON(config); err != nil {
		h.logger.Errorf("Failed to send STUN config: %v", err)
		return err
	}
	
	h.logger.Info("STUN config sent successfully")
	return nil
}

// GetConnectedUsers returns the number of connected users (for monitoring)
func (h *WebSocketHandler) GetConnectedUsers() int {
	// This would need to be implemented in the signaling service
	// For now, return 0 as placeholder
	return 0
}
