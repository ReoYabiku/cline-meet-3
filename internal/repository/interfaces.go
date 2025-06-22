package repository

import (
	"context"
	"github.com/signaling-server/internal/model"
)

// UserRepository defines the interface for user data operations
type User interface {
	SaveUser(ctx context.Context, user *model.UserSession) error
	GetUser(ctx context.Context, userID string) (*model.UserSession, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateUserRoom(ctx context.Context, userID, roomID string) error
}

// RoomRepository defines the interface for room data operations
type Room interface {
	SaveRoom(ctx context.Context, room *model.Room) error
	GetRoom(ctx context.Context, roomID string) (*model.Room, error)
	DeleteRoom(ctx context.Context, roomID string) error
	AddUserToRoom(ctx context.Context, roomID, userID string) error
	RemoveUserFromRoom(ctx context.Context, roomID, userID string) error
	GetRoomUsers(ctx context.Context, roomID string) ([]string, error)
}

// PubSubRepository defines the interface for pub/sub operations
type PubSub interface {
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
	Unsubscribe(ctx context.Context, channel string) error
}
