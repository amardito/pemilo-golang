package repository

import (
	"database/sql"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
)

type ticketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) domain.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *domain.Ticket) error {
	query := `
		INSERT INTO tickets (id, room_id, code, is_used, used_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, ticket.ID, ticket.RoomID, ticket.Code, ticket.IsUsed, ticket.UsedAt, ticket.CreatedAt, ticket.UpdatedAt)
	return err
}

func (r *ticketRepository) CreateBulk(tickets []*domain.Ticket) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO tickets (id, room_id, code, is_used, used_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ticket := range tickets {
		_, err = stmt.Exec(ticket.ID, ticket.RoomID, ticket.Code, ticket.IsUsed, ticket.UsedAt, ticket.CreatedAt, ticket.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ticketRepository) GetByCode(roomID string, code string) (*domain.Ticket, error) {
	query := `
		SELECT id, room_id, code, is_used, used_at, created_at, updated_at
		FROM tickets
		WHERE room_id = $1 AND code = $2
	`
	ticket := &domain.Ticket{}
	err := r.db.QueryRow(query, roomID, code).Scan(
		&ticket.ID,
		&ticket.RoomID,
		&ticket.Code,
		&ticket.IsUsed,
		&ticket.UsedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTicketNotFound
	}
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (r *ticketRepository) MarkAsUsed(id string) error {
	now := time.Now()
	query := `UPDATE tickets SET is_used = true, used_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(query, id, now, now)
	return err
}

func (r *ticketRepository) GetByRoomID(roomID string) ([]*domain.Ticket, error) {
	query := `
		SELECT id, room_id, code, is_used, used_at, created_at, updated_at
		FROM tickets
		WHERE room_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		ticket := &domain.Ticket{}
		err := rows.Scan(
			&ticket.ID,
			&ticket.RoomID,
			&ticket.Code,
			&ticket.IsUsed,
			&ticket.UsedAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (r *ticketRepository) Delete(id string) error {
	query := `DELETE FROM tickets WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ticketRepository) ExistsByCode(roomID string, code string) (bool, error) {
	query := `SELECT COUNT(*) FROM tickets WHERE room_id = $1 AND code = $2`
	var count int
	err := r.db.QueryRow(query, roomID, code).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ticketRepository) CountByRoomID(roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM tickets WHERE room_id = $1`
	var count int
	err := r.db.QueryRow(query, roomID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ticketRepository) CountUsedByRoomID(roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM tickets WHERE room_id = $1 AND is_used = true`
	var count int
	err := r.db.QueryRow(query, roomID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
