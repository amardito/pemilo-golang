package dto

import "time"

// LoginRequest represents admin login request with encrypted password
type LoginRequest struct {
	Username          string `json:"username" binding:"required"`
	EncryptedPassword string `json:"encrypted_password" binding:"required"`
}

// LoginResponse represents successful login response with JWT token
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Admin     AdminInfo `json:"admin"`
}

// AdminInfo represents admin information in responses
type AdminInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	MaxRoom   int    `json:"max_room"`
	MaxVoters int    `json:"max_voters"`
	IsActive  bool   `json:"is_active"`
}

// CreateAdminRequest represents request to create a new admin (Basic Auth protected)
// Password should be sent as plain text (protected by Basic Auth), not encrypted
type CreateAdminRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=8"`
	MaxRoom   int    `json:"max_room" binding:"required,min=1"`
	MaxVoters int    `json:"max_voters" binding:"required,min=1"`
}

// CreateAdminResponse represents admin creation response
type CreateAdminResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	MaxRoom   int       `json:"max_room"`
	MaxVoters int       `json:"max_voters"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAdminQuotaResponse represents admin quota usage
type GetAdminQuotaResponse struct {
	Admin         AdminInfo `json:"admin"`
	CurrentRooms  int       `json:"current_rooms"`
	CurrentVoters int       `json:"current_voters"`
	RoomLimit     int       `json:"room_limit"`
	VotersLimit   int       `json:"voters_limit"`
}
