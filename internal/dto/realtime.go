package dto

import "time"

// RealtimeVoteData represents real-time vote data for monitoring
type RealtimeVoteData struct {
	CandidateID   string    `json:"candidate_id"`
	CandidateName string    `json:"candidate_name"`
	VoteCount     int       `json:"vote_count"`
	Timestamp     time.Time `json:"timestamp"`
}

// RealtimeVoteResponse represents the real-time vote monitoring response
type RealtimeVoteResponse struct {
	RoomID     string              `json:"room_id"`
	RoomName   string              `json:"room_name"`
	VoteData   []*RealtimeVoteData `json:"vote_data"`
	TotalVotes int                 `json:"total_votes"`
	UpdatedAt  time.Time           `json:"updated_at"`
}
