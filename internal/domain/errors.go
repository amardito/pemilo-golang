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
	ErrTicketDuplicate   = errors.New("ticket code already exists in this room")
	ErrInvalidTicket     = errors.New("invalid ticket code")
	ErrTicketRequired    = errors.New("ticket code is required for this room")

	// Vote errors
	ErrVoterAlreadyVoted = errors.New("voter has already voted in this room")
	ErrVoteLimitReached  = errors.New("vote limit has been reached for this room")
	ErrInvalidVote       = errors.New("invalid vote data")

	// General errors
	ErrInvalidInput = errors.New("invalid input provided")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrForbidden    = errors.New("forbidden")

	// Admin/Auth errors
	ErrAdminNotFound      = errors.New("admin not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAdminExists        = errors.New("admin with this username already exists")
	ErrRateLimitExceeded  = errors.New("too many failed login attempts, please try again later")
	ErrQuotaExceeded      = errors.New("quota exceeded")
	ErrMaxRoomExceeded    = errors.New("maximum room limit exceeded for this admin")
	ErrMaxVotersExceeded  = errors.New("maximum voters limit exceeded for this admin")
	ErrAdminInactive      = errors.New("admin account is inactive")
)
