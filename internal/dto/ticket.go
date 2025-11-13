package dto

import "time"

// CreateTicketRequest represents the request to create a single ticket
type CreateTicketRequest struct {
	RoomID string `json:"room_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

// CreateTicketsBulkRequest represents the request to create tickets in bulk
type CreateTicketsBulkRequest struct {
	RoomID string   `json:"room_id" binding:"required"`
	Codes  []string `json:"codes" binding:"required,min=1"`
}

// TicketResponse represents a ticket response
type TicketResponse struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"room_id"`
	Code      string     `json:"code"`
	IsUsed    bool       `json:"is_used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ListTicketsResponse represents the list of tickets for a room
type ListTicketsResponse struct {
	Tickets []*TicketResponse `json:"tickets"`
}
