package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/signaling-server/internal/model"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

// User repository implementation
func (r *RedisRepository) SaveUser(ctx context.Context, user *model.UserSession) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	key := fmt.Sprintf("user:%s", user.ID)
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *RedisRepository) GetUser(ctx context.Context, userID string) (*model.UserSession, error) {
	key := fmt.Sprintf("user:%s", userID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user model.UserSession
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (r *RedisRepository) DeleteUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisRepository) UpdateUserRoom(ctx context.Context, userID, roomID string) error {
	user, err := r.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}

	user.RoomID = roomID
	user.LastSeen = time.Now()
	return r.SaveUser(ctx, user)
}

// Room repository implementation
func (r *RedisRepository) SaveRoom(ctx context.Context, room *model.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}

	key := fmt.Sprintf("room:%s", room.ID)
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *RedisRepository) GetRoom(ctx context.Context, roomID string) (*model.Room, error) {
	key := fmt.Sprintf("room:%s", roomID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	var room model.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}

	return &room, nil
}

func (r *RedisRepository) DeleteRoom(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("room:%s", roomID)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisRepository) AddUserToRoom(ctx context.Context, roomID, userID string) error {
	room, err := r.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		room = &model.Room{
			ID:        roomID,
			Users:     []string{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	if !room.AddUser(userID) {
		return fmt.Errorf("room is full or user already exists")
	}

	return r.SaveRoom(ctx, room)
}

func (r *RedisRepository) RemoveUserFromRoom(ctx context.Context, roomID, userID string) error {
	room, err := r.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return nil // Room doesn't exist, nothing to remove
	}

	room.RemoveUser(userID)

	if room.IsEmpty() {
		return r.DeleteRoom(ctx, roomID)
	}

	return r.SaveRoom(ctx, room)
}

func (r *RedisRepository) GetRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	room, err := r.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return []string{}, nil
	}

	return room.Users, nil
}

// PubSub repository implementation
func (r *RedisRepository) Publish(ctx context.Context, channel string, message []byte) error {
	return r.client.Publish(ctx, channel, message).Err()
}

func (r *RedisRepository) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	pubsub := r.client.Subscribe(ctx, channel)
	ch := pubsub.Channel()

	msgCh := make(chan []byte, 100)
	go func() {
		defer close(msgCh)
		for msg := range ch {
			msgCh <- []byte(msg.Payload)
		}
	}()

	return msgCh, nil
}

func (r *RedisRepository) Unsubscribe(ctx context.Context, channel string) error {
	// This is a simplified implementation
	// In a real scenario, you'd need to manage subscriptions more carefully
	return nil
}
