package domain

import "time"

// Candidate represents a candidate in an election
type Candidate struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	Name        string    `json:"name"`
	PhotoURL    string    `json:"photo_url"`
	Description string    `json:"description"` // Rich text
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SubCandidate represents a sub-candidate (e.g., vice president)
type SubCandidate struct {
	ID          string    `json:"id"`
	CandidateID string    `json:"candidate_id"`
	Name        string    `json:"name"`
	PhotoURL    string    `json:"photo_url"`
	Description string    `json:"description,omitempty"` // Optional
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CandidateWithSubs represents a candidate with their sub-candidates
type CandidateWithSubs struct {
	Candidate     Candidate      `json:"candidate"`
	SubCandidates []SubCandidate `json:"sub_candidates"`
}

// CandidateRepository defines the interface for candidate persistence
type CandidateRepository interface {
	Create(candidate *Candidate) error
	GetByID(id string) (*Candidate, error)
	GetByRoomID(roomID string) ([]*Candidate, error)
	Update(candidate *Candidate) error
	Delete(id string) error
}

// SubCandidateRepository defines the interface for sub-candidate persistence
type SubCandidateRepository interface {
	Create(subCandidate *SubCandidate) error
	GetByID(id string) (*SubCandidate, error)
	GetByCandidateID(candidateID string) ([]*SubCandidate, error)
	Update(subCandidate *SubCandidate) error
	Delete(id string) error
}
