package dto

import "time"

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	Name             string     `json:"name" binding:"required"`
	VotersType       string     `json:"voters_type" binding:"required,oneof=custom_tickets wild_limited wild_unlimited"`
	VotersLimit      *int       `json:"voters_limit,omitempty"`
	SessionStartTime *time.Time `json:"session_start_time,omitempty"`
	SessionEndTime   *time.Time `json:"session_end_time,omitempty"`
	Status           string     `json:"status" binding:"required,oneof=enabled disabled"`
	PublishState     string     `json:"publish_state" binding:"required,oneof=draft published"`
	// AdminID will be extracted from JWT token, not from request
}

// UpdateRoomRequest represents the request to update a room
type UpdateRoomRequest struct {
	Name             *string    `json:"name,omitempty"`
	VotersType       *string    `json:"voters_type,omitempty"`
	VotersLimit      *int       `json:"voters_limit,omitempty"`
	SessionStartTime *time.Time `json:"session_start_time,omitempty"`
	SessionEndTime   *time.Time `json:"session_end_time,omitempty"`
	Status           *string    `json:"status,omitempty"`
	PublishState     *string    `json:"publish_state,omitempty"`
}

// RoomResponse represents the room response
type RoomResponse struct {
	ID               string     `json:"id"`
	AdminID          string     `json:"admin_id"`
	Name             string     `json:"name"`
	VotersType       string     `json:"voters_type"`
	VotersLimit      *int       `json:"voters_limit,omitempty"`
	SessionStartTime *time.Time `json:"session_start_time,omitempty"`
	SessionEndTime   *time.Time `json:"session_end_time,omitempty"`
	Status           string     `json:"status"`
	PublishState     string     `json:"publish_state"`
	SessionState     string     `json:"session_state"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ListRoomsRequest represents filters for listing rooms
type ListRoomsRequest struct {
	Status       *string `form:"status"`
	PublishState *string `form:"publish_state"`
	SessionState *string `form:"session_state"`
}

// BulkDeleteRoomRequest represents the request to delete multiple rooms
type BulkDeleteRoomRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// ListRoomsResponse represents the list of rooms
type ListRoomsResponse struct {
	Rooms []*RoomResponse `json:"rooms"`
}
