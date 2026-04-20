package service

import (
	"context"
	"errors"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
	"github.com/amard/pemilo-golang/internal/repository"
)

var (
	ErrSlateNotFound    = errors.New("slate not found")
	ErrMaxSlatesReached = errors.New("maximum number of slates reached for this package")
	ErrSlateHasBallots  = errors.New("cannot delete slate with existing ballots")
	ErrSlateNotEditable = errors.New("slates can only be modified in DRAFT or SCHEDULED status")
	ErrMemberNotFound   = errors.New("slate member not found")
)

type SlateService struct {
	slateRepo *repository.SlateRepo
	eventRepo *repository.EventRepo
}

func NewSlateService(slateRepo *repository.SlateRepo, eventRepo *repository.EventRepo) *SlateService {
	return &SlateService{slateRepo: slateRepo, eventRepo: eventRepo}
}

func (s *SlateService) Create(ctx context.Context, eventID, userID string, req dto.CreateSlateRequest) (*model.Slate, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return nil, ErrEventLocked
	}

	count, err := s.slateRepo.CountByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if count >= event.MaxSlates {
		return nil, ErrMaxSlatesReached
	}

	return s.slateRepo.Create(ctx, eventID, req.Number, req.Name, req.Vision, req.Mission, req.PhotoURL)
}

func (s *SlateService) List(ctx context.Context, eventID string) ([]model.Slate, error) {
	slates, err := s.slateRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Load members for each slate
	for i := range slates {
		members, err := s.slateRepo.ListMembersBySlate(ctx, slates[i].ID)
		if err != nil {
			return nil, err
		}
		slates[i].Members = members
	}

	return slates, nil
}

func (s *SlateService) Update(ctx context.Context, slateID, userID string, req dto.UpdateSlateRequest) (*model.Slate, error) {
	slate, err := s.slateRepo.GetByID(ctx, slateID)
	if err != nil {
		return nil, ErrSlateNotFound
	}

	event, err := s.eventRepo.GetByID(ctx, slate.EventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return nil, ErrEventLocked
	}

	return s.slateRepo.Update(ctx, slateID, req.Number, req.Name, req.Vision, req.Mission, req.PhotoURL)
}

func (s *SlateService) Delete(ctx context.Context, slateID, userID string) error {
	slate, err := s.slateRepo.GetByID(ctx, slateID)
	if err != nil {
		return ErrSlateNotFound
	}

	event, err := s.eventRepo.GetByID(ctx, slate.EventID)
	if err != nil {
		return ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return ErrEventForbidden
	}

	// Only allow delete in DRAFT or SCHEDULED
	if event.Status != model.EventStatusDraft && event.Status != model.EventStatusScheduled {
		return ErrSlateNotEditable
	}

	hasBallots, err := s.slateRepo.HasBallots(ctx, slateID)
	if err != nil {
		return err
	}
	if hasBallots {
		return ErrSlateHasBallots
	}

	return s.slateRepo.Delete(ctx, slateID)
}

// ── Slate Members ──

func (s *SlateService) CreateMember(ctx context.Context, slateID, userID string, req dto.CreateSlateMemberRequest) (*model.SlateMember, error) {
	slate, err := s.slateRepo.GetByID(ctx, slateID)
	if err != nil {
		return nil, ErrSlateNotFound
	}

	event, err := s.eventRepo.GetByID(ctx, slate.EventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return nil, ErrEventLocked
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	return s.slateRepo.CreateMember(ctx, slateID, req.Role, req.FullName, req.PhotoURL, req.Bio, sortOrder)
}

func (s *SlateService) UpdateMember(ctx context.Context, memberID, userID string, req dto.UpdateSlateMemberRequest) (*model.SlateMember, error) {
	member, err := s.slateRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, ErrMemberNotFound
	}

	slate, err := s.slateRepo.GetByID(ctx, member.SlateID)
	if err != nil {
		return nil, ErrSlateNotFound
	}

	event, err := s.eventRepo.GetByID(ctx, slate.EventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return nil, ErrEventLocked
	}

	return s.slateRepo.UpdateMember(ctx, memberID, req.Role, req.FullName, req.PhotoURL, req.Bio, req.SortOrder)
}

func (s *SlateService) DeleteMember(ctx context.Context, memberID, userID string) error {
	member, err := s.slateRepo.GetMemberByID(ctx, memberID)
	if err != nil {
		return ErrMemberNotFound
	}

	slate, err := s.slateRepo.GetByID(ctx, member.SlateID)
	if err != nil {
		return ErrSlateNotFound
	}

	event, err := s.eventRepo.GetByID(ctx, slate.EventID)
	if err != nil {
		return ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return ErrEventLocked
	}

	return s.slateRepo.DeleteMember(ctx, memberID)
}
