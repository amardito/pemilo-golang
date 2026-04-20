package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
)

type AuditLogRepo struct {
	db *sql.DB
}

func NewAuditLogRepo(db *sql.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(ctx context.Context, eventID string, actorUserID *string, action string, meta string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (event_id, actor_user_id, action, meta) VALUES ($1, $2, $3, $4::jsonb)`,
		eventID, actorUserID, action, meta,
	)
	return err
}

func (r *AuditLogRepo) List(ctx context.Context, eventID string, page, perPage int) ([]model.AuditLog, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE event_id = $1`, eventID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, event_id, actor_user_id, action, meta, created_at
		 FROM audit_logs WHERE event_id = $1 ORDER BY created_at DESC LIMIT %d OFFSET %d`, perPage, offset),
		eventID,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.EventID, &l.ActorUserID, &l.Action, &l.Meta, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}

// ── Convenience method to convert to DTO ──

func AuditLogsToDTO(logs []model.AuditLog) []dto.AuditLogDTO {
	result := make([]dto.AuditLogDTO, len(logs))
	for i, l := range logs {
		result[i] = dto.AuditLogDTO{
			ID:          l.ID,
			Action:      l.Action,
			ActorUserID: l.ActorUserID,
			Meta:        l.Meta,
			CreatedAt:   l.CreatedAt,
		}
	}
	return result
}
