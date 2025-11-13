package domain

import "time"

// Vote represents a vote cast in an election
type Vote struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	CandidateID     string    `json:"candidate_id"`
	SubCandidateID  *string   `json:"sub_candidate_id,omitempty"`
	VoterIdentifier string    `json:"voter_identifier"` // ticket code or auto-generated ID
	CreatedAt       time.Time `json:"created_at"`
}

// VoteCount represents vote statistics for a candidate
type VoteCount struct {
	CandidateID string    `json:"candidate_id"`
	Count       int       `json:"count"`
	Timestamp   time.Time `json:"timestamp"`
}

// VoteRepository defines the interface for vote persistence
type VoteRepository interface {
	Create(vote *Vote) error
	GetByRoomID(roomID string) ([]*Vote, error)
	GetVoteCountByRoom(roomID string) ([]*VoteCount, error)
	GetTotalVoteCountByRoom(roomID string) (int, error)
	CheckVoterHasVoted(roomID string, voterIdentifier string) (bool, error)
	GetRealtimeVoteCounts(roomID string) ([]*VoteCount, error)
}
