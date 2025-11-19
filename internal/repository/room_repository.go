package repository

import (
	"database/sql"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
)

type roomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) domain.RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, admin_id, name, voters_type, voters_limit, session_start_time, session_end_time, status, publish_state, session_state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(query, room.ID, room.AdminID, room.Name, room.VotersType, room.VotersLimit, room.SessionStartTime, room.SessionEndTime, room.Status, room.PublishState, room.SessionState, room.CreatedAt, room.UpdatedAt)
	return err
}

func (r *roomRepository) GetByID(id string) (*domain.Room, error) {
	query := `
		SELECT id, admin_id, name, voters_type, voters_limit, session_start_time, session_end_time, status, publish_state, session_state, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`
	room := &domain.Room{}
	err := r.db.QueryRow(query, id).Scan(
		&room.ID,
		&room.AdminID,
		&room.Name,
		&room.VotersType,
		&room.VotersLimit,
		&room.SessionStartTime,
		&room.SessionEndTime,
		&room.Status,
		&room.PublishState,
		&room.SessionState,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *roomRepository) Update(room *domain.Room) error {
	query := `
		UPDATE rooms
		SET admin_id = $2, name = $3, voters_type = $4, voters_limit = $5, session_start_time = $6, session_end_time = $7, status = $8, publish_state = $9, session_state = $10, updated_at = $11
		WHERE id = $1
	`
	room.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, room.ID, room.AdminID, room.Name, room.VotersType, room.VotersLimit, room.SessionStartTime, room.SessionEndTime, room.Status, room.PublishState, room.SessionState, room.UpdatedAt)
	return err
}

func (r *roomRepository) Delete(id string) error {
	query := `DELETE FROM rooms WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *roomRepository) List(filters domain.RoomFilters) ([]*domain.Room, error) {
	query := `
		SELECT id, admin_id, name, voters_type, voters_limit, session_start_time, session_end_time, status, publish_state, session_state, created_at, updated_at
		FROM rooms
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filters.Status != nil {
		query += ` AND status = $` + string(rune(argCount+'0'))
		args = append(args, *filters.Status)
		argCount++
	}

	if filters.PublishState != nil {
		query += ` AND publish_state = $` + string(rune(argCount+'0'))
		args = append(args, *filters.PublishState)
		argCount++
	}

	if filters.SessionState != nil {
		query += ` AND session_state = $` + string(rune(argCount+'0'))
		args = append(args, *filters.SessionState)
		argCount++
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		err := rows.Scan(
			&room.ID,
			&room.AdminID,
			&room.Name,
			&room.VotersType,
			&room.VotersLimit,
			&room.SessionStartTime,
			&room.SessionEndTime,
			&room.Status,
			&room.PublishState,
			&room.SessionState,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (r *roomRepository) UpdateSessionState(roomID string, state domain.SessionState) error {
	query := `UPDATE rooms SET session_state = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(query, roomID, state, time.Now())
	return err
}

func (r *roomRepository) CountByAdminID(adminID string) (int, error) {
	query := `SELECT COUNT(*) FROM rooms WHERE admin_id = $1`
	var count int
	err := r.db.QueryRow(query, adminID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
