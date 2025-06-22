package service

import (
	"context"
	"fmt"

	"github.com/signaling-server/internal/model"
	"github.com/signaling-server/internal/repository"
)

type RoomService struct {
	roomRepo repository.Room
	userRepo repository.User
}

func NewRoomService(roomRepo repository.Room, userRepo repository.User) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

// JoinRoom adds a user to a room
func (s *RoomService) JoinRoom(ctx context.Context, userID, roomID string) (*model.Room, error) {
	// Check if room exists and has space
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// If room doesn't exist, it will be created in AddUserToRoom
	if room != nil && !room.CanJoin() {
		return nil, fmt.Errorf("room is full")
	}

	// Add user to room
	if err := s.roomRepo.AddUserToRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	// Update user's room
	if err := s.userRepo.UpdateUserRoom(ctx, userID, roomID); err != nil {
		return nil, err
	}

	// Return updated room
	return s.roomRepo.GetRoom(ctx, roomID)
}

// LeaveRoom removes a user from a room
func (s *RoomService) LeaveRoom(ctx context.Context, userID, roomID string) error {
	// Remove user from room
	if err := s.roomRepo.RemoveUserFromRoom(ctx, roomID, userID); err != nil {
		return err
	}

	// Update user's room to empty
	if err := s.userRepo.UpdateUserRoom(ctx, userID, ""); err != nil {
		return err
	}

	return nil
}

// GetRoom retrieves a room by ID
func (s *RoomService) GetRoom(ctx context.Context, roomID string) (*model.Room, error) {
	return s.roomRepo.GetRoom(ctx, roomID)
}

// GetRoomUsers retrieves all users in a room
func (s *RoomService) GetRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	return s.roomRepo.GetRoomUsers(ctx, roomID)
}

// GetOtherUsersInRoom returns all users in the room except the specified user
func (s *RoomService) GetOtherUsersInRoom(ctx context.Context, roomID, excludeUserID string) ([]string, error) {
	users, err := s.roomRepo.GetRoomUsers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	var otherUsers []string
	for _, userID := range users {
		if userID != excludeUserID {
			otherUsers = append(otherUsers, userID)
		}
	}

	return otherUsers, nil
}

// IsRoomFull checks if a room is at capacity
func (s *RoomService) IsRoomFull(ctx context.Context, roomID string) (bool, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomID)
	if err != nil {
		return false, err
	}
	if room == nil {
		return false, nil // Room doesn't exist, so not full
	}

	return !room.CanJoin(), nil
}
