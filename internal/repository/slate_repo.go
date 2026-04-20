package repository

import (
	"context"
	"database/sql"

	"github.com/amard/pemilo-golang/internal/model"
)

type SlateRepo struct {
	db *sql.DB
}

func NewSlateRepo(db *sql.DB) *SlateRepo {
	return &SlateRepo{db: db}
}

func (r *SlateRepo) Create(ctx context.Context, eventID string, number int, name string, vision, mission, photoURL *string) (*model.Slate, error) {
	var s model.Slate
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO slates (event_id, number, name, vision, mission, photo_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, event_id, number, name, vision, mission, photo_url, created_at`,
		eventID, number, name, vision, mission, photoURL,
	).Scan(&s.ID, &s.EventID, &s.Number, &s.Name, &s.Vision, &s.Mission, &s.PhotoURL, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SlateRepo) ListByEvent(ctx context.Context, eventID string) ([]model.Slate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, event_id, number, name, vision, mission, photo_url, created_at
		 FROM slates WHERE event_id = $1 ORDER BY number`, eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slates []model.Slate
	for rows.Next() {
		var s model.Slate
		if err := rows.Scan(&s.ID, &s.EventID, &s.Number, &s.Name, &s.Vision, &s.Mission, &s.PhotoURL, &s.CreatedAt); err != nil {
			return nil, err
		}
		slates = append(slates, s)
	}
	return slates, rows.Err()
}

func (r *SlateRepo) GetByID(ctx context.Context, id string) (*model.Slate, error) {
	var s model.Slate
	err := r.db.QueryRowContext(ctx,
		`SELECT id, event_id, number, name, vision, mission, photo_url, created_at
		 FROM slates WHERE id = $1`, id,
	).Scan(&s.ID, &s.EventID, &s.Number, &s.Name, &s.Vision, &s.Mission, &s.PhotoURL, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SlateRepo) Update(ctx context.Context, id string, number *int, name *string, vision, mission, photoURL *string) (*model.Slate, error) {
	var s model.Slate
	err := r.db.QueryRowContext(ctx,
		`UPDATE slates SET
			number = COALESCE($2, number),
			name = COALESCE($3, name),
			vision = COALESCE($4, vision),
			mission = COALESCE($5, mission),
			photo_url = COALESCE($6, photo_url)
		 WHERE id = $1
		 RETURNING id, event_id, number, name, vision, mission, photo_url, created_at`,
		id, number, name, vision, mission, photoURL,
	).Scan(&s.ID, &s.EventID, &s.Number, &s.Name, &s.Vision, &s.Mission, &s.PhotoURL, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SlateRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM slates WHERE id = $1`, id)
	return err
}

func (r *SlateRepo) CountByEvent(ctx context.Context, eventID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM slates WHERE event_id = $1`, eventID).Scan(&count)
	return count, err
}

func (r *SlateRepo) HasBallots(ctx context.Context, slateID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ballots WHERE slate_id = $1`, slateID).Scan(&count)
	return count > 0, err
}

// ── Slate Members ──

func (r *SlateRepo) CreateMember(ctx context.Context, slateID, role, fullName string, photoURL, bio *string, sortOrder int) (*model.SlateMember, error) {
	var m model.SlateMember
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO slate_members (slate_id, role, full_name, photo_url, bio, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, slate_id, role, full_name, photo_url, bio, sort_order`,
		slateID, role, fullName, photoURL, bio, sortOrder,
	).Scan(&m.ID, &m.SlateID, &m.Role, &m.FullName, &m.PhotoURL, &m.Bio, &m.SortOrder)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SlateRepo) ListMembersBySlate(ctx context.Context, slateID string) ([]model.SlateMember, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slate_id, role, full_name, photo_url, bio, sort_order
		 FROM slate_members WHERE slate_id = $1 ORDER BY sort_order`, slateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.SlateMember
	for rows.Next() {
		var m model.SlateMember
		if err := rows.Scan(&m.ID, &m.SlateID, &m.Role, &m.FullName, &m.PhotoURL, &m.Bio, &m.SortOrder); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *SlateRepo) GetMemberByID(ctx context.Context, id string) (*model.SlateMember, error) {
	var m model.SlateMember
	err := r.db.QueryRowContext(ctx,
		`SELECT id, slate_id, role, full_name, photo_url, bio, sort_order
		 FROM slate_members WHERE id = $1`, id,
	).Scan(&m.ID, &m.SlateID, &m.Role, &m.FullName, &m.PhotoURL, &m.Bio, &m.SortOrder)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SlateRepo) UpdateMember(ctx context.Context, id string, role, fullName *string, photoURL, bio *string, sortOrder *int) (*model.SlateMember, error) {
	var m model.SlateMember
	err := r.db.QueryRowContext(ctx,
		`UPDATE slate_members SET
			role = COALESCE($2, role),
			full_name = COALESCE($3, full_name),
			photo_url = COALESCE($4, photo_url),
			bio = COALESCE($5, bio),
			sort_order = COALESCE($6, sort_order)
		 WHERE id = $1
		 RETURNING id, slate_id, role, full_name, photo_url, bio, sort_order`,
		id, role, fullName, photoURL, bio, sortOrder,
	).Scan(&m.ID, &m.SlateID, &m.Role, &m.FullName, &m.PhotoURL, &m.Bio, &m.SortOrder)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SlateRepo) DeleteMember(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM slate_members WHERE id = $1`, id)
	return err
}
