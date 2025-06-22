package model

import (
	"time"

	"github.com/gorilla/websocket"
)

// User represents a connected user
type User struct {
	ID         string          `json:"id"`
	SessionID  string          `json:"session_id"`
	RoomID     string          `json:"room_id,omitempty"`
	Connection *websocket.Conn `json:"-"`
	CreatedAt  time.Time       `json:"created_at"`
	LastSeen   time.Time       `json:"last_seen"`
}

// UserSession represents user session data stored in Redis
type UserSession struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	RoomID    string    `json:"room_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// ToSession converts User to UserSession for Redis storage
func (u *User) ToSession() *UserSession {
	return &UserSession{
		ID:        u.ID,
		SessionID: u.SessionID,
		RoomID:    u.RoomID,
		CreatedAt: u.CreatedAt,
		LastSeen:  u.LastSeen,
	}
}
