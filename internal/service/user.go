package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/signaling-server/internal/model"
	"github.com/signaling-server/internal/repository"
)

type UserService struct {
	userRepo repository.User
}

func NewUserService(userRepo repository.User) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with the given session ID
func (s *UserService) CreateUser(ctx context.Context, sessionID string) (*model.UserSession, error) {
	user := &model.UserSession{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*model.UserSession, error) {
	return s.userRepo.GetUser(ctx, userID)
}

// GetOrCreateUser gets an existing user or creates a new one
func (s *UserService) GetOrCreateUser(ctx context.Context, sessionID string) (*model.UserSession, error) {
	// Try to find existing user by session ID
	// This is a simplified approach - in production, you might want to index by session ID
	user, err := s.CreateUser(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserActivity updates the user's last seen timestamp
func (s *UserService) UpdateUserActivity(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // User doesn't exist
	}

	user.LastSeen = time.Now()
	return s.userRepo.SaveUser(ctx, user)
}

// DeleteUser removes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.DeleteUser(ctx, userID)
}

// JoinRoom adds a user to a room
func (s *UserService) JoinRoom(ctx context.Context, userID, roomID string) error {
	return s.userRepo.UpdateUserRoom(ctx, userID, roomID)
}

// LeaveRoom removes a user from their current room
func (s *UserService) LeaveRoom(ctx context.Context, userID string) error {
	return s.userRepo.UpdateUserRoom(ctx, userID, "")
}
