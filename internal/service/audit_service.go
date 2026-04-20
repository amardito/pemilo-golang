package service

import (
	"context"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/repository"
)

type AuditService struct {
	auditLogRepo *repository.AuditLogRepo
	eventRepo    *repository.EventRepo
}

func NewAuditService(auditLogRepo *repository.AuditLogRepo, eventRepo *repository.EventRepo) *AuditService {
	return &AuditService{auditLogRepo: auditLogRepo, eventRepo: eventRepo}
}

func (s *AuditService) List(ctx context.Context, eventID, userID string, page, perPage int) (*dto.AuditLogListResponse, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	logs, total, err := s.auditLogRepo.List(ctx, eventID, page, perPage)
	if err != nil {
		return nil, err
	}

	return &dto.AuditLogListResponse{
		Logs:    repository.AuditLogsToDTO(logs),
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}
