package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
	"github.com/amard/pemilo-golang/internal/repository"
	"github.com/amard/pemilo-golang/internal/util"
)

var (
	ErrEventNotOpen     = errors.New("voting is not open")
	ErrInvalidToken     = errors.New("invalid token or NIM")
	ErrAlreadyVoted     = errors.New("you have already voted")
	ErrVoterNotEligible = errors.New("voter is not eligible")
	ErrInvalidSlate     = errors.New("invalid slate selection")
)

type VoteService struct {
	db             *sql.DB
	eventRepo      *repository.EventRepo
	slateRepo      *repository.SlateRepo
	voterRepo      *repository.VoterRepo
	voterTokenRepo *repository.VoterTokenRepo
	ballotRepo     *repository.BallotRepo
}

func NewVoteService(
	db *sql.DB,
	eventRepo *repository.EventRepo,
	slateRepo *repository.SlateRepo,
	voterRepo *repository.VoterRepo,
	voterTokenRepo *repository.VoterTokenRepo,
	ballotRepo *repository.BallotRepo,
) *VoteService {
	return &VoteService{
		db:             db,
		eventRepo:      eventRepo,
		slateRepo:      slateRepo,
		voterRepo:      voterRepo,
		voterTokenRepo: voterTokenRepo,
		ballotRepo:     ballotRepo,
	}
}

func (s *VoteService) Prepare(ctx context.Context, eventID string, req dto.VotePrepareRequest) (*dto.VotePrepareResponse, error) {
	// Validate event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}

	if !s.isEventOpen(event) {
		return nil, ErrEventNotOpen
	}

	// Normalize inputs
	token := util.NormalizeToken(req.Token)
	if !util.ValidateToken(token) {
		return nil, ErrInvalidToken
	}
	nimNorm := util.NormalizeNIM(req.NIM)

	// Find token
	vt, err := s.voterTokenRepo.GetByEventAndToken(ctx, eventID, token)
	if err != nil {
		return nil, ErrInvalidToken // generic error
	}

	if vt.Status != model.TokenStatusActive {
		return nil, ErrAlreadyVoted
	}

	// Find voter and verify NIM match
	voter, err := s.voterRepo.GetByEventAndNIM(ctx, eventID, nimNorm)
	if err != nil || voter.ID != vt.VoterID {
		return nil, ErrInvalidToken // generic error
	}

	if voter.Status != model.VoterStatusEligible {
		return nil, ErrVoterNotEligible
	}

	if voter.HasVoted {
		return nil, ErrAlreadyVoted
	}

	// Fetch slates with members
	slates, err := s.slateRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	slatesPublic := make([]dto.SlatePublic, len(slates))
	for i, sl := range slates {
		members, err := s.slateRepo.ListMembersBySlate(ctx, sl.ID)
		if err != nil {
			return nil, err
		}

		membersPublic := make([]dto.SlateMemberPublic, len(members))
		for j, m := range members {
			membersPublic[j] = dto.SlateMemberPublic{
				Role:      m.Role,
				FullName:  m.FullName,
				PhotoURL:  m.PhotoURL,
				Bio:       m.Bio,
				SortOrder: m.SortOrder,
			}
		}

		slatesPublic[i] = dto.SlatePublic{
			ID:       sl.ID,
			Number:   sl.Number,
			Name:     sl.Name,
			Vision:   sl.Vision,
			Mission:  sl.Mission,
			PhotoURL: sl.PhotoURL,
			Members:  membersPublic,
		}
	}

	return &dto.VotePrepareResponse{
		OK: true,
		VoterDisplay: dto.VoterDisplay{
			FullName:  voter.FullName,
			ClassName: voter.ClassName,
		},
		Slates:    slatesPublic,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

func (s *VoteService) Submit(ctx context.Context, eventID string, req dto.VoteSubmitRequest) error {
	// Validate event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return ErrEventNotFound
	}
	if !s.isEventOpen(event) {
		return ErrEventNotOpen
	}

	// Normalize inputs
	token := util.NormalizeToken(req.Token)
	if !util.ValidateToken(token) {
		return ErrInvalidToken
	}
	nimNorm := util.NormalizeNIM(req.NIM)

	// Verify slate exists for this event
	slate, err := s.slateRepo.GetByID(ctx, req.SlateID)
	if err != nil || slate.EventID != eventID {
		return ErrInvalidSlate
	}

	// ── ATOMIC TRANSACTION ──
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1) Lock token row FOR UPDATE
	vt, err := s.voterTokenRepo.LockForUpdate(ctx, tx, eventID, token)
	if err != nil {
		return ErrInvalidToken
	}

	// 2) Verify token is ACTIVE
	if vt.Status != model.TokenStatusActive {
		return ErrAlreadyVoted
	}

	// 3) Fetch voter via token's voter_id, verify NIM match
	voter, err := s.voterRepo.GetByEventAndNIM(ctx, eventID, nimNorm)
	if err != nil || voter.ID != vt.VoterID {
		return ErrInvalidToken
	}

	if voter.HasVoted {
		return ErrAlreadyVoted
	}

	// 4) Insert ballot — NO voter_id (secret ballot)
	if err := s.ballotRepo.InsertInTx(ctx, tx, eventID, req.SlateID); err != nil {
		return err
	}

	// 5) Mark voter as voted (with guard)
	rowsAffected, err := s.voterRepo.MarkVoted(ctx, tx, voter.ID)
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrAlreadyVoted
	}

	// 6) Mark token as USED
	if err := s.voterTokenRepo.MarkUsed(ctx, tx, vt.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *VoteService) isEventOpen(event *model.Event) bool {
	if event.Status != model.EventStatusOpen {
		return false
	}

	now := time.Now()
	if event.OpensAt != nil && now.Before(*event.OpensAt) {
		return false
	}
	if event.ClosesAt != nil && now.After(*event.ClosesAt) {
		return false
	}

	return true
}
