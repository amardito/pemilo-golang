package service

import (
	"context"
	"errors"
	"io"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
	"github.com/amard/pemilo-golang/internal/repository"
	"github.com/amard/pemilo-golang/internal/util"
)

var (
	ErrMaxVotersReached = errors.New("maximum number of voters reached for this package")
)

type VoterService struct {
	voterRepo      *repository.VoterRepo
	voterTokenRepo *repository.VoterTokenRepo
	eventRepo      *repository.EventRepo
	auditLogRepo   *repository.AuditLogRepo
}

func NewVoterService(
	voterRepo *repository.VoterRepo,
	voterTokenRepo *repository.VoterTokenRepo,
	eventRepo *repository.EventRepo,
	auditLogRepo *repository.AuditLogRepo,
) *VoterService {
	return &VoterService{
		voterRepo:      voterRepo,
		voterTokenRepo: voterTokenRepo,
		eventRepo:      eventRepo,
		auditLogRepo:   auditLogRepo,
	}
}

func (s *VoterService) ImportCSV(ctx context.Context, eventID, userID string, file io.Reader) (*dto.ImportResult, error) {
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

	// Check current count
	currentCount, err := s.voterRepo.CountByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	parsed, err := util.ParseVotersCSV(file)
	if err != nil {
		return nil, err
	}

	// Check limit
	if currentCount+len(parsed.Rows) > event.MaxVoters {
		return nil, ErrMaxVotersReached
	}

	// Convert to bulk insert format
	bulkRows := make([]struct {
		FullName      string
		NIMRaw        string
		NIMNormalized string
		ClassName     string
	}, len(parsed.Rows))

	for i, row := range parsed.Rows {
		bulkRows[i] = struct {
			FullName      string
			NIMRaw        string
			NIMNormalized string
			ClassName     string
		}{
			FullName:      row.FullName,
			NIMRaw:        row.NIMRaw,
			NIMNormalized: row.NIMNormalized,
			ClassName:     row.ClassName,
		}
	}

	imported, dbRejected, err := s.voterRepo.BulkInsert(ctx, eventID, bulkRows)
	if err != nil {
		return nil, err
	}

	// Merge CSV-level rejections with DB-level rejections
	allRejected := make([]dto.ImportReject, 0, len(parsed.Rejected)+len(dbRejected))
	for _, r := range parsed.Rejected {
		allRejected = append(allRejected, dto.ImportReject{Row: r.Row, Reason: r.Reason})
	}
	allRejected = append(allRejected, dbRejected...)

	s.auditLogRepo.Create(ctx, eventID, &userID, "voters.imported", `{}`)

	return &dto.ImportResult{
		ImportedCount: imported,
		Rejected:      allRejected,
	}, nil
}

func (s *VoterService) List(ctx context.Context, eventID, userID string, params dto.VoterListParams) (*dto.VoterListResponse, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	voters, total, err := s.voterRepo.List(ctx, eventID, params)
	if err != nil {
		return nil, err
	}

	voterDTOs := make([]dto.VoterDTO, len(voters))
	for i, v := range voters {
		voterDTOs[i] = dto.VoterDTO{
			ID:        v.ID,
			FullName:  v.FullName,
			NIMRaw:    v.NIMRaw,
			ClassName: v.ClassName,
			HasVoted:  v.HasVoted,
			VotedAt:   v.VotedAt,
			Status:    string(v.Status),
		}
	}

	return &dto.VoterListResponse{
		Voters:  voterDTOs,
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
	}, nil
}

func (s *VoterService) GenerateTokens(ctx context.Context, eventID, userID string) (int, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return 0, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return 0, ErrEventForbidden
	}
	if event.Status == model.EventStatusLocked {
		return 0, ErrEventLocked
	}

	voters, err := s.voterRepo.GetVotersWithoutToken(ctx, eventID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, voter := range voters {
		token, err := util.GenerateToken()
		if err != nil {
			return count, err
		}

		_, err = s.voterTokenRepo.Create(ctx, eventID, voter.ID, token)
		if err != nil {
			// Collision — retry once
			token, err = util.GenerateToken()
			if err != nil {
				return count, err
			}
			_, err = s.voterTokenRepo.Create(ctx, eventID, voter.ID, token)
			if err != nil {
				continue
			}
		}
		count++
	}

	s.auditLogRepo.Create(ctx, eventID, &userID, "tokens.generated", `{}`)
	return count, nil
}

func (s *VoterService) ExportTokens(ctx context.Context, eventID, userID string) ([]repository.TokenExportRow, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	return s.voterTokenRepo.GetTokensForExport(ctx, eventID)
}

func (s *VoterService) ExportTurnout(ctx context.Context, eventID, userID string) ([]model.Voter, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	return s.voterRepo.GetAllVotersForExport(ctx, eventID)
}
