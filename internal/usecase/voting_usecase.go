package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/google/uuid"
)

// VotingUsecase handles all voter-related business logic
type VotingUsecase struct {
	voteRepo      domain.VoteRepository
	roomRepo      domain.RoomRepository
	candidateRepo domain.CandidateRepository
	ticketRepo    domain.TicketRepository
}

func NewVotingUsecase(
	voteRepo domain.VoteRepository,
	roomRepo domain.RoomRepository,
	candidateRepo domain.CandidateRepository,
	ticketRepo domain.TicketRepository,
) *VotingUsecase {
	return &VotingUsecase{
		voteRepo:      voteRepo,
		roomRepo:      roomRepo,
		candidateRepo: candidateRepo,
		ticketRepo:    ticketRepo,
	}
}

// GetVoterRoomInfo returns room information and validates voter eligibility
func (u *VotingUsecase) GetVoterRoomInfo(roomID string) (*domain.Room, []*domain.CandidateWithSubs, bool, string, error) {
	// Get room
	room, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, nil, false, "", err
	}

	// Check room is enabled
	if room.Status == domain.RoomStatusDisabled {
		return room, nil, false, "Room is disabled", domain.ErrRoomDisabled
	}

	// Check room is published
	if room.PublishState == domain.PublishStateDraft {
		return room, nil, false, "Room is not published", domain.ErrRoomNotPublished
	}

	// Check session state
	if room.SessionState == domain.SessionStateClosed {
		return room, nil, false, "Voting session is closed", domain.ErrSessionClosed
	}

	// For wild_unlimited, check time range
	if room.VotersType == domain.VotersTypeWildUnlimited {
		if !room.IsSessionActive() {
			return room, nil, false, "Voting session is not active", domain.ErrSessionNotActive
		}
	}

	// Get candidates
	candidates, err := u.getCandidatesWithSubs(roomID)
	if err != nil {
		return room, nil, false, "", err
	}

	requiresTicket := room.VotersType == domain.VotersTypeCustomTickets
	return room, candidates, requiresTicket, "", nil
}

// CastVote handles the complete voting flow with validation based on voters_type
func (u *VotingUsecase) CastVote(roomID, candidateID string, subCandidateID *string, ticketCode *string) (*domain.Vote, error) {
	// Get room
	room, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, err
	}

	// Validate room state
	if room.Status == domain.RoomStatusDisabled {
		return nil, domain.ErrRoomDisabled
	}
	if room.PublishState == domain.PublishStateDraft {
		return nil, domain.ErrRoomNotPublished
	}
	if room.SessionState == domain.SessionStateClosed {
		return nil, domain.ErrSessionClosed
	}

	// Verify candidate exists
	_, err = u.candidateRepo.GetByID(candidateID)
	if err != nil {
		return nil, err
	}

	var voterIdentifier string

	// Handle voting based on voters_type
	switch room.VotersType {
	case domain.VotersTypeCustomTickets:
		// Ticket is required
		if ticketCode == nil || *ticketCode == "" {
			return nil, domain.ErrTicketRequired
		}

		// Verify ticket
		ticket, err := u.ticketRepo.GetByCode(roomID, *ticketCode)
		if err != nil {
			return nil, domain.ErrInvalidTicket
		}

		if ticket.IsUsed {
			return nil, domain.ErrTicketAlreadyUsed
		}

		voterIdentifier = *ticketCode

		// Check if voter already voted
		hasVoted, err := u.voteRepo.CheckVoterHasVoted(roomID, voterIdentifier)
		if err != nil {
			return nil, err
		}
		if hasVoted {
			return nil, domain.ErrVoterAlreadyVoted
		}

		// Mark ticket as used
		if err := u.ticketRepo.MarkAsUsed(ticket.ID); err != nil {
			return nil, err
		}

	case domain.VotersTypeWildLimited:
		// Check if vote limit reached
		totalVotes, err := u.voteRepo.GetTotalVoteCountByRoom(roomID)
		if err != nil {
			return nil, err
		}

		if room.VotersLimit != nil && totalVotes >= *room.VotersLimit {
			// Close session automatically
			_ = u.roomRepo.UpdateSessionState(roomID, domain.SessionStateClosed)
			return nil, domain.ErrVoteLimitReached
		}

		// Generate unique voter identifier
		voterIdentifier = generateVoterID()

	case domain.VotersTypeWildUnlimited:
		// Check session active time range
		if !room.IsSessionActive() {
			return nil, domain.ErrSessionNotActive
		}

		// Generate unique voter identifier
		voterIdentifier = generateVoterID()

	default:
		return nil, domain.ErrInvalidVotersType
	}

	// Create vote
	vote := &domain.Vote{
		ID:              uuid.New().String(),
		RoomID:          roomID,
		CandidateID:     candidateID,
		SubCandidateID:  subCandidateID,
		VoterIdentifier: voterIdentifier,
		CreatedAt:       time.Now(),
	}

	if err := u.voteRepo.Create(vote); err != nil {
		return nil, err
	}

	// For wild_limited, check if limit reached after this vote
	if room.VotersType == domain.VotersTypeWildLimited {
		totalVotes, _ := u.voteRepo.GetTotalVoteCountByRoom(roomID)
		if room.VotersLimit != nil && totalVotes >= *room.VotersLimit {
			_ = u.roomRepo.UpdateSessionState(roomID, domain.SessionStateClosed)
		}
	}

	return vote, nil
}

// GetRealtimeVoteData returns real-time vote statistics for monitoring
func (u *VotingUsecase) GetRealtimeVoteData(roomID string) (*domain.Room, []*domain.VoteCount, int, error) {
	// Get room
	room, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get vote counts
	voteCounts, err := u.voteRepo.GetRealtimeVoteCounts(roomID)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get total votes
	totalVotes, err := u.voteRepo.GetTotalVoteCountByRoom(roomID)
	if err != nil {
		return nil, nil, 0, err
	}

	return room, voteCounts, totalVotes, nil
}

// Helper functions
func (u *VotingUsecase) getCandidatesWithSubs(roomID string) ([]*domain.CandidateWithSubs, error) {
	candidates, err := u.candidateRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}

	var result []*domain.CandidateWithSubs
	for _, candidate := range candidates {
		result = append(result, &domain.CandidateWithSubs{
			Candidate:     *candidate,
			SubCandidates: []domain.SubCandidate{},
		})
	}

	return result, nil
}

// generateVoterID creates a unique identifier for wild voters
func generateVoterID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("voter_%s", hex.EncodeToString(b))
}
