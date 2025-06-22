package model

import "encoding/json"

// MessageType represents the type of WebRTC signaling message
type MessageType string

const (
	MessageTypeJoinRoom     MessageType = "join_room"
	MessageTypeLeaveRoom    MessageType = "leave_room"
	MessageTypeOffer        MessageType = "offer"
	MessageTypeAnswer       MessageType = "answer"
	MessageTypeIceCandidate MessageType = "ice_candidate"
	MessageTypeUserJoined   MessageType = "user_joined"
	MessageTypeUserLeft     MessageType = "user_left"
	MessageTypeRoomFull     MessageType = "room_full"
	MessageTypeError        MessageType = "error"
)

// Message represents a WebRTC signaling message
type Message struct {
	Type      MessageType     `json:"type"`
	RoomID    string          `json:"room_id,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	TargetID  string          `json:"target_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// OfferData represents WebRTC offer data
type OfferData struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// AnswerData represents WebRTC answer data
type AnswerData struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// IceCandidateData represents WebRTC ICE candidate data
type IceCandidateData struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// JoinRoomData represents join room request data
type JoinRoomData struct {
	RoomID string `json:"room_id"`
}

// ErrorData represents error message data
type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// UserJoinedData represents user joined notification data
type UserJoinedData struct {
	UserID string   `json:"user_id"`
	Users  []string `json:"users"`
}

// UserLeftData represents user left notification data
type UserLeftData struct {
	UserID string   `json:"user_id"`
	Users  []string `json:"users"`
}
