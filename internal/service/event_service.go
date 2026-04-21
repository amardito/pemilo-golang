package service

import (
	"context"
	"errors"
	"time"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
	"github.com/amard/pemilo-golang/internal/repository"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrEventForbidden    = errors.New("forbidden: not event owner")
	ErrEventLocked       = errors.New("event is locked, no modifications allowed")
	ErrInvalidTransition = errors.New("invalid status transition")
)

type EventService struct {
	eventRepo    *repository.EventRepo
	auditLogRepo *repository.AuditLogRepo
}

func NewEventService(eventRepo *repository.EventRepo, auditLogRepo *repository.AuditLogRepo) *EventService {
	return &EventService{eventRepo: eventRepo, auditLogRepo: auditLogRepo}
}

func (s *EventService) Create(ctx context.Context, ownerID string, req dto.CreateEventRequest) (*model.Event, error) {
	limits := model.PackageLimitsMap[model.PackageFree]

	var opensAt, closesAt *string
	if req.OpensAt != nil {
		v := req.OpensAt.Format(time.RFC3339)
		opensAt = &v
	}
	if req.ClosesAt != nil {
		v := req.ClosesAt.Format(time.RFC3339)
		closesAt = &v
	}

	event, err := s.eventRepo.Create(ctx, ownerID, req.Title, req.Description, opensAt, closesAt, limits.MaxSlates, limits.MaxVoters, string(model.PackageFree))
	if err != nil {
		return nil, err
	}

	s.auditLogRepo.Create(ctx, event.ID, &ownerID, "event.created", `{}`)
	return event, nil
}

func (s *EventService) GetByID(ctx context.Context, eventID, userID string) (*model.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}
	return event, nil
}

// GetPublicInfo returns limited event info without ownership check — safe for public pages.
func (s *EventService) GetPublicInfo(ctx context.Context, eventID string) (*dto.EventPublicInfo, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	return &dto.EventPublicInfo{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Status:      string(event.Status),
		OpensAt:     event.OpensAt,
		ClosesAt:    event.ClosesAt,
	}, nil
}

func (s *EventService) List(ctx context.Context, userID string) ([]model.Event, error) {
	return s.eventRepo.ListByOwner(ctx, userID)
}

func (s *EventService) Update(ctx context.Context, eventID, userID string, req dto.UpdateEventRequest) (*model.Event, error) {
	event, err := s.GetByID(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if event.Status == model.EventStatusLocked {
		return nil, ErrEventLocked
	}

	var opensAt, closesAt *string
	if req.OpensAt != nil {
		v := req.OpensAt.Format(time.RFC3339)
		opensAt = &v
	}
	if req.ClosesAt != nil {
		v := req.ClosesAt.Format(time.RFC3339)
		closesAt = &v
	}

	updated, err := s.eventRepo.Update(ctx, eventID, req.Title, req.Description, opensAt, closesAt)
	if err != nil {
		return nil, err
	}

	s.auditLogRepo.Create(ctx, eventID, &userID, "event.updated", `{}`)
	return updated, nil
}

func (s *EventService) Open(ctx context.Context, eventID, userID string) error {
	event, err := s.GetByID(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if event.Status != model.EventStatusDraft && event.Status != model.EventStatusScheduled && event.Status != model.EventStatusClosed {
		return ErrInvalidTransition
	}
	if err := s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusOpen); err != nil {
		return err
	}
	s.auditLogRepo.Create(ctx, eventID, &userID, "event.opened", `{}`)
	return nil
}

func (s *EventService) Close(ctx context.Context, eventID, userID string) error {
	event, err := s.GetByID(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if event.Status != model.EventStatusOpen {
		return ErrInvalidTransition
	}
	if err := s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusClosed); err != nil {
		return err
	}
	s.auditLogRepo.Create(ctx, eventID, &userID, "event.closed", `{}`)
	return nil
}

func (s *EventService) Lock(ctx context.Context, eventID, userID string) error {
	event, err := s.GetByID(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if event.Status != model.EventStatusClosed {
		return ErrInvalidTransition
	}
	if err := s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusLocked); err != nil {
		return err
	}
	s.auditLogRepo.Create(ctx, eventID, &userID, "event.locked", `{}`)
	return nil
}

// GetByIDPublic returns event without ownership check (for public endpoints).
func (s *EventService) GetByIDPublic(ctx context.Context, eventID string) (*model.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}
