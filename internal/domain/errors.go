package domain

import "errors"

// Domain errors
var (
	// Room errors
	ErrRoomNotFound         = errors.New("room not found")
	ErrInvalidRoomName      = errors.New("room name is required")
	ErrInvalidVotersType    = errors.New("invalid voters type")
	ErrVotersLimitRequired  = errors.New("voters limit is required for wild_limited type")
	ErrSessionRangeRequired = errors.New("session time range is required for wild_unlimited type")
	ErrInvalidSessionRange  = errors.New("session end time must be after start time")
	ErrRoomDisabled         = errors.New("room is disabled")
	ErrRoomNotPublished     = errors.New("room is not published")
	ErrSessionClosed        = errors.New("voting session is closed")
	ErrSessionNotActive     = errors.New("voting session is not active")

	// Candidate errors
	ErrCandidateNotFound = errors.New("candidate not found")
	ErrInvalidCandidate  = errors.New("invalid candidate data")

	// Ticket errors
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketAlreadyUsed = errors.New("ticket has already been used")
	ErrInvalidTicket     = errors.New("invalid ticket code")
	ErrTicketRequired    = errors.New("ticket code is required for this room")

	// Vote errors
	ErrVoterAlreadyVoted = errors.New("voter has already voted in this room")
	ErrVoteLimitReached  = errors.New("vote limit has been reached for this room")
	ErrInvalidVote       = errors.New("invalid vote data")

	// General errors
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("forbidden")
)
