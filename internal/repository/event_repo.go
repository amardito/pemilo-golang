package repository

import (
	"context"
	"database/sql"

	"github.com/amard/pemilo-golang/internal/model"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Create(ctx context.Context, ownerID, title string, description *string, opensAt, closesAt *string, maxSlates, maxVoters int, pkg string) (*model.Event, error) {
	var e model.Event
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO events (owner_user_id, title, description, opens_at, closes_at, max_slates, max_voters, package)
		 VALUES ($1, $2, $3, $4::timestamptz, $5::timestamptz, $6, $7, $8)
		 RETURNING id, owner_user_id, title, description, status, opens_at, closes_at, max_slates, max_voters, package, created_at, updated_at`,
		ownerID, title, description, opensAt, closesAt, maxSlates, maxVoters, pkg,
	).Scan(&e.ID, &e.OwnerUserID, &e.Title, &e.Description, &e.Status, &e.OpensAt, &e.ClosesAt, &e.MaxSlates, &e.MaxVoters, &e.Package, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventRepo) GetByID(ctx context.Context, id string) (*model.Event, error) {
	var e model.Event
	err := r.db.QueryRowContext(ctx,
		`SELECT id, owner_user_id, title, description, status, opens_at, closes_at, max_slates, max_voters, package, created_at, updated_at
		 FROM events WHERE id = $1`, id,
	).Scan(&e.ID, &e.OwnerUserID, &e.Title, &e.Description, &e.Status, &e.OpensAt, &e.ClosesAt, &e.MaxSlates, &e.MaxVoters, &e.Package, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventRepo) ListByOwner(ctx context.Context, ownerID string) ([]model.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_user_id, title, description, status, opens_at, closes_at, max_slates, max_voters, package, created_at, updated_at
		 FROM events WHERE owner_user_id = $1 ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.OwnerUserID, &e.Title, &e.Description, &e.Status, &e.OpensAt, &e.ClosesAt, &e.MaxSlates, &e.MaxVoters, &e.Package, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepo) Update(ctx context.Context, id string, title *string, description *string, opensAt, closesAt *string) (*model.Event, error) {
	var e model.Event
	err := r.db.QueryRowContext(ctx,
		`UPDATE events SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			opens_at = COALESCE($4::timestamptz, opens_at),
			closes_at = COALESCE($5::timestamptz, closes_at),
			updated_at = now()
		 WHERE id = $1
		 RETURNING id, owner_user_id, title, description, status, opens_at, closes_at, max_slates, max_voters, package, created_at, updated_at`,
		id, title, description, opensAt, closesAt,
	).Scan(&e.ID, &e.OwnerUserID, &e.Title, &e.Description, &e.Status, &e.OpensAt, &e.ClosesAt, &e.MaxSlates, &e.MaxVoters, &e.Package, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventRepo) UpdateStatus(ctx context.Context, id string, status model.EventStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE events SET status = $2, updated_at = now() WHERE id = $1`,
		id, string(status),
	)
	return err
}

func (r *EventRepo) UpdatePackage(ctx context.Context, id string, pkg model.Package, maxSlates, maxVoters int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE events SET package = $2, max_slates = $3, max_voters = $4, updated_at = now() WHERE id = $1`,
		id, string(pkg), maxSlates, maxVoters,
	)
	return err
}
