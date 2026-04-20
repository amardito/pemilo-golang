package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
)

type VoterRepo struct {
	db *sql.DB
}

func NewVoterRepo(db *sql.DB) *VoterRepo {
	return &VoterRepo{db: db}
}

func (r *VoterRepo) BulkInsert(ctx context.Context, eventID string, rows []struct {
	FullName      string
	NIMRaw        string
	NIMNormalized string
	ClassName     string
}) (int, []dto.ImportReject, error) {
	imported := 0
	var rejected []dto.ImportReject

	for i, row := range rows {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO voters (event_id, full_name, nim_raw, nim_normalized, class_name)
			 VALUES ($1, $2, $3, $4, NULLIF($5, ''))`,
			eventID, row.FullName, row.NIMRaw, row.NIMNormalized, row.ClassName,
		)
		if err != nil {
			if strings.Contains(err.Error(), "uq_voters_event_nim") {
				rejected = append(rejected, dto.ImportReject{Row: i + 2, Reason: "duplicate nim in event"})
			} else {
				rejected = append(rejected, dto.ImportReject{Row: i + 2, Reason: err.Error()})
			}
			continue
		}
		imported++
	}
	return imported, rejected, nil
}

func (r *VoterRepo) CountByEvent(ctx context.Context, eventID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voters WHERE event_id = $1`, eventID).Scan(&count)
	return count, err
}

func (r *VoterRepo) List(ctx context.Context, eventID string, params dto.VoterListParams) ([]model.Voter, int, error) {
	where := []string{"event_id = $1"}
	args := []interface{}{eventID}
	argIdx := 2

	if params.Query != "" {
		where = append(where, fmt.Sprintf("(full_name ILIKE $%d OR nim_raw ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+params.Query+"%")
		argIdx++
	}
	if params.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.HasVoted != nil {
		where = append(where, fmt.Sprintf("has_voted = $%d", argIdx))
		args = append(args, *params.HasVoted)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voters WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Page
	limit := params.PerPage
	if limit <= 0 {
		limit = 20
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(
		`SELECT id, event_id, full_name, nim_raw, nim_normalized, class_name, status, has_voted, voted_at, created_at
		 FROM voters WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d`,
		whereClause, limit, offset,
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var voters []model.Voter
	for rows.Next() {
		var v model.Voter
		if err := rows.Scan(&v.ID, &v.EventID, &v.FullName, &v.NIMRaw, &v.NIMNormalized, &v.ClassName, &v.Status, &v.HasVoted, &v.VotedAt, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		voters = append(voters, v)
	}
	return voters, total, rows.Err()
}

func (r *VoterRepo) GetByEventAndNIM(ctx context.Context, eventID, nimNormalized string) (*model.Voter, error) {
	var v model.Voter
	err := r.db.QueryRowContext(ctx,
		`SELECT id, event_id, full_name, nim_raw, nim_normalized, class_name, status, has_voted, voted_at, created_at
		 FROM voters WHERE event_id = $1 AND nim_normalized = $2`,
		eventID, nimNormalized,
	).Scan(&v.ID, &v.EventID, &v.FullName, &v.NIMRaw, &v.NIMNormalized, &v.ClassName, &v.Status, &v.HasVoted, &v.VotedAt, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VoterRepo) MarkVoted(ctx context.Context, tx *sql.Tx, voterID string) (int64, error) {
	result, err := tx.ExecContext(ctx,
		`UPDATE voters SET has_voted = true, voted_at = now() WHERE id = $1 AND has_voted = false`,
		voterID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetVotersWithoutToken returns voters that don't have an ACTIVE token.
func (r *VoterRepo) GetVotersWithoutToken(ctx context.Context, eventID string) ([]model.Voter, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT v.id, v.event_id, v.full_name, v.nim_raw, v.nim_normalized, v.class_name, v.status, v.has_voted, v.voted_at, v.created_at
		 FROM voters v
		 LEFT JOIN voter_tokens vt ON v.id = vt.voter_id AND vt.status = 'ACTIVE'
		 WHERE v.event_id = $1 AND v.status = 'ELIGIBLE' AND vt.id IS NULL`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voters []model.Voter
	for rows.Next() {
		var v model.Voter
		if err := rows.Scan(&v.ID, &v.EventID, &v.FullName, &v.NIMRaw, &v.NIMNormalized, &v.ClassName, &v.Status, &v.HasVoted, &v.VotedAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		voters = append(voters, v)
	}
	return voters, rows.Err()
}

// GetAllVotersForExport returns all voters for turnout export.
func (r *VoterRepo) GetAllVotersForExport(ctx context.Context, eventID string) ([]model.Voter, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, event_id, full_name, nim_raw, nim_normalized, class_name, status, has_voted, voted_at, created_at
		 FROM voters WHERE event_id = $1 ORDER BY full_name`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voters []model.Voter
	for rows.Next() {
		var v model.Voter
		if err := rows.Scan(&v.ID, &v.EventID, &v.FullName, &v.NIMRaw, &v.NIMNormalized, &v.ClassName, &v.Status, &v.HasVoted, &v.VotedAt, &v.CreatedAt); err != nil {
			return nil, err
		}
		voters = append(voters, v)
	}
	return voters, rows.Err()
}
