package dto

import "time"

// VoteRequest represents the request to cast a vote
type VoteRequest struct {
	RoomID         string  `json:"room_id" binding:"required"`
	CandidateID    string  `json:"candidate_id" binding:"required"`
	SubCandidateID *string `json:"sub_candidate_id,omitempty"`
	TicketCode     *string `json:"ticket_code,omitempty"` // Required for custom_tickets rooms
}

// VoteResponse represents a vote response
type VoteResponse struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	CandidateID     string    `json:"candidate_id"`
	SubCandidateID  *string   `json:"sub_candidate_id,omitempty"`
	VoterIdentifier string    `json:"voter_identifier"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetVoterRoomInfoResponse represents the response for voter room information
type GetVoterRoomInfoResponse struct {
	Room           RoomResponse         `json:"room"`
	Candidates     []*CandidateResponse `json:"candidates"`
	RequiresTicket bool                 `json:"requires_ticket"`
	IsActive       bool                 `json:"is_active"`
	Message        string               `json:"message,omitempty"`
}

// VerifyTicketRequest represents the request to verify a ticket
type VerifyTicketRequest struct {
	RoomID     string `json:"room_id" binding:"required"`
	TicketCode string `json:"ticket_code" binding:"required"`
}

// VerifyTicketResponse represents the response for ticket verification
type VerifyTicketResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}
