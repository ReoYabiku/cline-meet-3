package model

import (
	"time"
)

const MaxRoomUsers = 10

// Room represents a signaling room
type Room struct {
	ID        string    `json:"id"`
	Users     []string  `json:"users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanJoin checks if a user can join the room
func (r *Room) CanJoin() bool {
	return len(r.Users) < MaxRoomUsers
}

// AddUser adds a user to the room
func (r *Room) AddUser(userID string) bool {
	if !r.CanJoin() {
		return false
	}
	
	// Check if user is already in the room
	for _, id := range r.Users {
		if id == userID {
			return true // User already in room
		}
	}
	
	r.Users = append(r.Users, userID)
	r.UpdatedAt = time.Now()
	return true
}

// RemoveUser removes a user from the room
func (r *Room) RemoveUser(userID string) bool {
	for i, id := range r.Users {
		if id == userID {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			r.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// IsEmpty checks if the room is empty
func (r *Room) IsEmpty() bool {
	return len(r.Users) == 0
}

// GetOtherUsers returns all users except the specified one
func (r *Room) GetOtherUsers(userID string) []string {
	var others []string
	for _, id := range r.Users {
		if id != userID {
			others = append(others, id)
		}
	}
	return others
}
