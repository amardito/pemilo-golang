package repository

import (
	"context"
	"database/sql"

	"github.com/amard/pemilo-golang/internal/model"
)

type VoterTokenRepo struct {
	db *sql.DB
}

func NewVoterTokenRepo(db *sql.DB) *VoterTokenRepo {
	return &VoterTokenRepo{db: db}
}

func (r *VoterTokenRepo) Create(ctx context.Context, eventID, voterID, token string) (*model.VoterToken, error) {
	var vt model.VoterToken
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO voter_tokens (event_id, voter_id, token) VALUES ($1, $2, $3)
		 RETURNING id, event_id, voter_id, token, status, issued_at, used_at`,
		eventID, voterID, token,
	).Scan(&vt.ID, &vt.EventID, &vt.VoterID, &vt.Token, &vt.Status, &vt.IssuedAt, &vt.UsedAt)
	if err != nil {
		return nil, err
	}
	return &vt, nil
}

func (r *VoterTokenRepo) GetByEventAndToken(ctx context.Context, eventID, token string) (*model.VoterToken, error) {
	var vt model.VoterToken
	err := r.db.QueryRowContext(ctx,
		`SELECT id, event_id, voter_id, token, status, issued_at, used_at
		 FROM voter_tokens WHERE event_id = $1 AND token = $2`,
		eventID, token,
	).Scan(&vt.ID, &vt.EventID, &vt.VoterID, &vt.Token, &vt.Status, &vt.IssuedAt, &vt.UsedAt)
	if err != nil {
		return nil, err
	}
	return &vt, nil
}

// LockForUpdate acquires a row lock on the token within a transaction.
func (r *VoterTokenRepo) LockForUpdate(ctx context.Context, tx *sql.Tx, eventID, token string) (*model.VoterToken, error) {
	var vt model.VoterToken
	err := tx.QueryRowContext(ctx,
		`SELECT id, event_id, voter_id, token, status, issued_at, used_at
		 FROM voter_tokens WHERE event_id = $1 AND token = $2 FOR UPDATE`,
		eventID, token,
	).Scan(&vt.ID, &vt.EventID, &vt.VoterID, &vt.Token, &vt.Status, &vt.IssuedAt, &vt.UsedAt)
	if err != nil {
		return nil, err
	}
	return &vt, nil
}

// MarkUsed sets token status to USED within a transaction.
func (r *VoterTokenRepo) MarkUsed(ctx context.Context, tx *sql.Tx, tokenID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE voter_tokens SET status = 'USED', used_at = now() WHERE id = $1`,
		tokenID,
	)
	return err
}

// GetTokensForExport returns tokens with voter info for CSV export.
type TokenExportRow struct {
	FullName  string
	NIMRaw    string
	ClassName *string
	Token     string
}

func (r *VoterTokenRepo) GetTokensForExport(ctx context.Context, eventID string) ([]TokenExportRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT v.full_name, v.nim_raw, v.class_name, vt.token
		 FROM voter_tokens vt
		 JOIN voters v ON v.id = vt.voter_id
		 WHERE vt.event_id = $1
		 ORDER BY v.full_name`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TokenExportRow
	for rows.Next() {
		var row TokenExportRow
		if err := rows.Scan(&row.FullName, &row.NIMRaw, &row.ClassName, &row.Token); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// GetTokenMapByEventID returns a map of voter_id -> token for all ACTIVE tokens in an event.
func (r *VoterTokenRepo) GetTokenMapByEventID(ctx context.Context, eventID string) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT voter_id, token FROM voter_tokens WHERE event_id = $1 AND status = 'ACTIVE'`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var voterID, token string
		if err := rows.Scan(&voterID, &token); err != nil {
			return nil, err
		}
		m[voterID] = token
	}
	return m, rows.Err()
}
