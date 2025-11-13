package dto

import "time"

// CreateCandidateRequest represents the request to create a candidate
type CreateCandidateRequest struct {
	RoomID        string                      `json:"room_id" binding:"required"`
	Name          string                      `json:"name" binding:"required"`
	PhotoURL      string                      `json:"photo_url" binding:"required"`
	Description   string                      `json:"description"`
	SubCandidates []CreateSubCandidateRequest `json:"sub_candidates,omitempty"`
}

// CreateSubCandidateRequest represents the request to create a sub-candidate
type CreateSubCandidateRequest struct {
	Name        string `json:"name" binding:"required"`
	PhotoURL    string `json:"photo_url" binding:"required"`
	Description string `json:"description,omitempty"`
}

// UpdateCandidateRequest represents the request to update a candidate
type UpdateCandidateRequest struct {
	Name        *string `json:"name,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateSubCandidateRequest represents the request to update a sub-candidate
type UpdateSubCandidateRequest struct {
	Name        *string `json:"name,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CandidateResponse represents a candidate response
type CandidateResponse struct {
	ID            string                 `json:"id"`
	RoomID        string                 `json:"room_id"`
	Name          string                 `json:"name"`
	PhotoURL      string                 `json:"photo_url"`
	Description   string                 `json:"description"`
	SubCandidates []SubCandidateResponse `json:"sub_candidates,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// SubCandidateResponse represents a sub-candidate response
type SubCandidateResponse struct {
	ID          string    `json:"id"`
	CandidateID string    `json:"candidate_id"`
	Name        string    `json:"name"`
	PhotoURL    string    `json:"photo_url"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListCandidatesResponse represents the list of candidates for a room
type ListCandidatesResponse struct {
	Candidates []*CandidateResponse `json:"candidates"`
}
