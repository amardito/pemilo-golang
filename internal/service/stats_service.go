package service

import (
	"context"
	"time"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/repository"
)

type StatsService struct {
	ballotRepo *repository.BallotRepo
	eventRepo  *repository.EventRepo
}

func NewStatsService(ballotRepo *repository.BallotRepo, eventRepo *repository.EventRepo) *StatsService {
	return &StatsService{ballotRepo: ballotRepo, eventRepo: eventRepo}
}

func (s *StatsService) GetStats(ctx context.Context, eventID, userID string) (*dto.StatsResponse, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	total, voted, err := s.ballotRepo.GetTurnoutCounts(ctx, eventID)
	if err != nil {
		return nil, err
	}

	votesBySlate, err := s.ballotRepo.GetVotesBySlate(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if votesBySlate == nil {
		votesBySlate = []dto.SlateVotes{}
	}

	latestVoters, err := s.ballotRepo.GetLatestVoters(ctx, eventID, 10)
	if err != nil {
		return nil, err
	}
	if latestVoters == nil {
		latestVoters = []dto.LatestVoter{}
	}

	return &dto.StatsResponse{
		EventID:       eventID,
		TotalVoters:   total,
		VotedCount:    voted,
		NotVotedCount: total - voted,
		VotesBySlate:  votesBySlate,
		LatestVoters:  latestVoters,
		UpdatedAt:     time.Now(),
	}, nil
}
