package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/signaling-server/internal/model"
	"github.com/signaling-server/internal/repository"
	"github.com/signaling-server/pkg/logger"
)

type SignalingService struct {
	userService *UserService
	roomService *RoomService
	pubsub      repository.PubSub
	logger      *logger.Logger
	
	// Connection management
	connections map[string]*model.User
	connMutex   sync.RWMutex
}

func NewSignalingService(
	userService *UserService,
	roomService *RoomService,
	pubsub repository.PubSub,
	logger *logger.Logger,
) *SignalingService {
	return &SignalingService{
		userService: userService,
		roomService: roomService,
		pubsub:      pubsub,
		logger:      logger,
		connections: make(map[string]*model.User),
	}
}

// AddConnection adds a WebSocket connection
func (s *SignalingService) AddConnection(userID string, conn *websocket.Conn, sessionID string) (*model.User, error) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	user := &model.User{
		ID:         userID,
		SessionID:  sessionID,
		Connection: conn,
		CreatedAt:  time.Now(),
		LastSeen:   time.Now(),
	}

	s.connections[userID] = user
	s.logger.Infof("User connected: %s", userID)

	return user, nil
}

// RemoveConnection removes a WebSocket connection
func (s *SignalingService) RemoveConnection(userID string) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if user, exists := s.connections[userID]; exists {
		// Leave room if user is in one
		if user.RoomID != "" {
			ctx := context.Background()
			s.handleLeaveRoom(ctx, user, user.RoomID)
		}
		
		delete(s.connections, userID)
		s.logger.Infof("User disconnected: %s", userID)
	}
}

// GetConnection retrieves a WebSocket connection
func (s *SignalingService) GetConnection(userID string) (*model.User, bool) {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()
	
	user, exists := s.connections[userID]
	return user, exists
}

// HandleMessage processes incoming WebSocket messages
func (s *SignalingService) HandleMessage(ctx context.Context, userID string, messageData []byte) error {
	var msg model.Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	msg.UserID = userID
	msg.Timestamp = time.Now().Unix()

	user, exists := s.GetConnection(userID)
	if !exists {
		return fmt.Errorf("user connection not found: %s", userID)
	}

	switch msg.Type {
	case model.MessageTypeJoinRoom:
		return s.handleJoinRoom(ctx, user, &msg)
	case model.MessageTypeLeaveRoom:
		return s.handleLeaveRoom(ctx, user, msg.RoomID)
	case model.MessageTypeOffer:
		return s.handleOffer(ctx, user, &msg)
	case model.MessageTypeAnswer:
		return s.handleAnswer(ctx, user, &msg)
	case model.MessageTypeIceCandidate:
		return s.handleIceCandidate(ctx, user, &msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleJoinRoom processes join room requests
func (s *SignalingService) handleJoinRoom(ctx context.Context, user *model.User, msg *model.Message) error {
	s.logger.Infof("Received join room message: %s", string(msg.Data))
	
	var joinData model.JoinRoomData
	if err := json.Unmarshal(msg.Data, &joinData); err != nil {
		s.logger.Errorf("Failed to unmarshal join room data: %v, raw data: %s", err, string(msg.Data))
		return s.sendError(user, 400, "Invalid join room data")
	}
	
	s.logger.Infof("Parsed join room data: %+v", joinData)

	// If user is already in a room, leave it first
	if user.RoomID != "" && user.RoomID != joinData.RoomID {
		s.logger.Infof("User %s is already in room %s, leaving before joining %s", user.ID, user.RoomID, joinData.RoomID)
		s.handleLeaveRoom(ctx, user, user.RoomID)
	}

	// Clean up disconnected users from the room before checking if it's full
	if err := s.cleanupDisconnectedUsersFromRoom(ctx, joinData.RoomID); err != nil {
		s.logger.Errorf("Failed to cleanup disconnected users from room %s: %v", joinData.RoomID, err)
	}

	// Check if room is full
	isFull, err := s.roomService.IsRoomFull(ctx, joinData.RoomID)
	if err != nil {
		return s.sendError(user, 500, "Failed to check room status")
	}
	
	// Log room status for debugging
	roomUsers, _ := s.roomService.GetRoomUsers(ctx, joinData.RoomID)
	s.logger.Infof("Room %s status: isFull=%v, current users=%v", joinData.RoomID, isFull, roomUsers)
	
	if isFull {
		return s.sendMessage(user, &model.Message{
			Type:      model.MessageTypeRoomFull,
			RoomID:    joinData.RoomID,
			Timestamp: time.Now().Unix(),
		})
	}

	// Join room
	_, err = s.roomService.JoinRoom(ctx, user.ID, joinData.RoomID)
	if err != nil {
		s.logger.Errorf("Failed to join room %s for user %s: %v", joinData.RoomID, user.ID, err)
		return s.sendError(user, 500, "Failed to join room")
	}

	// Update user's room
	s.connMutex.Lock()
	user.RoomID = joinData.RoomID
	s.connMutex.Unlock()

	s.logger.Infof("User %s successfully joined room %s", user.ID, joinData.RoomID)

	// Get other users in the room and filter for only connected users
	otherUsers, err := s.roomService.GetOtherUsersInRoom(ctx, joinData.RoomID, user.ID)
	if err != nil {
		s.logger.Errorf("Failed to get other users: %v", err)
	}
	
	// Filter out disconnected users
	connectedUsers := s.filterConnectedUsers(otherUsers)
	activeUsers := append(connectedUsers, user.ID) // Include the joining user
	
	s.logger.Infof("Room %s now has %d active users: %v", joinData.RoomID, len(activeUsers), activeUsers)
	
	if len(connectedUsers) > 0 {
		userJoinedMsg := &model.Message{
			Type:      model.MessageTypeUserJoined,
			RoomID:    joinData.RoomID,
			UserID:    user.ID,
			Timestamp: time.Now().Unix(),
		}
		
		userData := model.UserJoinedData{
			UserID: user.ID,
			Users:  activeUsers,
		}
		userJoinedMsg.Data, _ = json.Marshal(userData)

		s.broadcastToUsers(connectedUsers, userJoinedMsg)
		s.logger.Infof("Notified %d connected users about new user %s joining room %s", len(connectedUsers), user.ID, joinData.RoomID)
	}

	// Send confirmation to joining user with only connected users
	return s.sendMessage(user, &model.Message{
		Type:      model.MessageTypeUserJoined,
		RoomID:    joinData.RoomID,
		UserID:    user.ID,
		Timestamp: time.Now().Unix(),
		Data:      func() json.RawMessage { d, _ := json.Marshal(model.UserJoinedData{UserID: user.ID, Users: activeUsers}); return d }(),
	})
}

// handleLeaveRoom processes leave room requests
func (s *SignalingService) handleLeaveRoom(ctx context.Context, user *model.User, roomID string) error {
	if user.RoomID == "" {
		return nil // User not in a room
	}

	// Get other users before leaving
	otherUsers, err := s.roomService.GetOtherUsersInRoom(ctx, user.RoomID, user.ID)
	if err != nil {
		s.logger.Errorf("Failed to get other users: %v", err)
	}

	// Leave room
	if err := s.roomService.LeaveRoom(ctx, user.ID, user.RoomID); err != nil {
		return s.sendError(user, 500, "Failed to leave room")
	}

	oldRoomID := user.RoomID

	// Update user's room
	s.connMutex.Lock()
	user.RoomID = ""
	s.connMutex.Unlock()

	// Notify other users
	if len(otherUsers) > 0 {
		userLeftMsg := &model.Message{
			Type:      model.MessageTypeUserLeft,
			RoomID:    oldRoomID,
			UserID:    user.ID,
			Timestamp: time.Now().Unix(),
		}
		
		userData := model.UserLeftData{
			UserID: user.ID,
			Users:  otherUsers,
		}
		userLeftMsg.Data, _ = json.Marshal(userData)

		s.broadcastToUsers(otherUsers, userLeftMsg)
	}

	return nil
}

// handleOffer processes WebRTC offer messages
func (s *SignalingService) handleOffer(ctx context.Context, user *model.User, msg *model.Message) error {
	if user.RoomID == "" {
		return s.sendError(user, 400, "User not in a room")
	}

	s.logger.Infof("Handling offer from user %s to target %s", user.ID, msg.TargetID)

	// Forward offer to target user
	if msg.TargetID != "" {
		// Set the sender's user ID in the message
		msg.UserID = user.ID
		return s.forwardToUser(msg.TargetID, msg)
	}

	return s.sendError(user, 400, "Target user ID required for offer")
}

// handleAnswer processes WebRTC answer messages
func (s *SignalingService) handleAnswer(ctx context.Context, user *model.User, msg *model.Message) error {
	if user.RoomID == "" {
		return s.sendError(user, 400, "User not in a room")
	}

	s.logger.Infof("Handling answer from user %s to target %s", user.ID, msg.TargetID)

	// Forward answer to target user
	if msg.TargetID != "" {
		// Set the sender's user ID in the message
		msg.UserID = user.ID
		return s.forwardToUser(msg.TargetID, msg)
	}

	return s.sendError(user, 400, "Target user ID required for answer")
}

// handleIceCandidate processes ICE candidate messages
func (s *SignalingService) handleIceCandidate(ctx context.Context, user *model.User, msg *model.Message) error {
	if user.RoomID == "" {
		return s.sendError(user, 400, "User not in a room")
	}

	s.logger.Infof("Handling ICE candidate from user %s to target %s", user.ID, msg.TargetID)

	// Forward ICE candidate to target user
	if msg.TargetID != "" {
		// Set the sender's user ID in the message
		msg.UserID = user.ID
		return s.forwardToUser(msg.TargetID, msg)
	}

	return s.sendError(user, 400, "Target user ID required for ICE candidate")
}

// Helper methods
func (s *SignalingService) sendMessage(user *model.User, msg *model.Message) error {
	s.logger.Infof("Sending message to user %s: type=%s", user.ID, msg.Type)
	if err := user.Connection.WriteJSON(msg); err != nil {
		s.logger.Errorf("Failed to send message to user %s: %v", user.ID, err)
		return err
	}
	s.logger.Infof("Message sent successfully to user %s", user.ID)
	return nil
}

func (s *SignalingService) sendError(user *model.User, code int, message string) error {
	errorMsg := &model.Message{
		Type:      model.MessageTypeError,
		Timestamp: time.Now().Unix(),
	}
	
	errorData := model.ErrorData{
		Code:    code,
		Message: message,
	}
	errorMsg.Data, _ = json.Marshal(errorData)

	return s.sendMessage(user, errorMsg)
}

func (s *SignalingService) forwardToUser(targetUserID string, msg *model.Message) error {
	targetUser, exists := s.GetConnection(targetUserID)
	if !exists {
		return fmt.Errorf("target user not connected: %s", targetUserID)
	}

	return s.sendMessage(targetUser, msg)
}

func (s *SignalingService) broadcastToUsers(userIDs []string, msg *model.Message) {
	for _, userID := range userIDs {
		if user, exists := s.GetConnection(userID); exists {
			if err := s.sendMessage(user, msg); err != nil {
				s.logger.Errorf("Failed to send message to user %s: %v", userID, err)
			}
		}
	}
}

// filterConnectedUsers filters a list of user IDs to only include those with active connections
func (s *SignalingService) filterConnectedUsers(userIDs []string) []string {
	var connectedUsers []string
	for _, userID := range userIDs {
		if _, exists := s.GetConnection(userID); exists {
			connectedUsers = append(connectedUsers, userID)
		}
	}
	return connectedUsers
}

// cleanupDisconnectedUsersFromRoom removes disconnected users from a room
func (s *SignalingService) cleanupDisconnectedUsersFromRoom(ctx context.Context, roomID string) error {
	roomUsers, err := s.roomService.GetRoomUsers(ctx, roomID)
	if err != nil {
		return err
	}

	var disconnectedUsers []string
	for _, userID := range roomUsers {
		if _, exists := s.GetConnection(userID); !exists {
			disconnectedUsers = append(disconnectedUsers, userID)
		}
	}

	// Remove disconnected users from the room
	for _, userID := range disconnectedUsers {
		s.logger.Infof("Removing disconnected user %s from room %s", userID, roomID)
		if err := s.roomService.LeaveRoom(ctx, userID, roomID); err != nil {
			s.logger.Errorf("Failed to remove disconnected user %s from room %s: %v", userID, roomID, err)
		}
	}

	if len(disconnectedUsers) > 0 {
		s.logger.Infof("Cleaned up %d disconnected users from room %s", len(disconnectedUsers), roomID)
	}

	return nil
}
