package repository

import (
	"context"
	"database/sql"

	"github.com/amard/pemilo-golang/internal/model"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, eventID string, pkg model.Package, amount int) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO orders (event_id, package, amount) VALUES ($1, $2, $3)
		 RETURNING id, event_id, package, amount, status, ipaymu_reference, created_at, updated_at`,
		eventID, string(pkg), amount,
	).Scan(&o.ID, &o.EventID, &o.Package, &o.Amount, &o.Status, &o.IPaymuReference, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, event_id, package, amount, status, ipaymu_reference, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.EventID, &o.Package, &o.Amount, &o.Status, &o.IPaymuReference, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetByReference(ctx context.Context, ref string) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, event_id, package, amount, status, ipaymu_reference, created_at, updated_at
		 FROM orders WHERE ipaymu_reference = $1`, ref,
	).Scan(&o.ID, &o.EventID, &o.Package, &o.Amount, &o.Status, &o.IPaymuReference, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) UpdateReference(ctx context.Context, id, reference string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET ipaymu_reference = $2, updated_at = now() WHERE id = $1`,
		id, reference,
	)
	return err
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $2, updated_at = now() WHERE id = $1`,
		id, string(status),
	)
	return err
}
