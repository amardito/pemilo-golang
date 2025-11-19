package repository

import (
	"database/sql"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
)

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(admin *domain.Admin) error {
	query := `
		INSERT INTO admins (id, username, password, max_room, max_voters, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, admin.ID, admin.Username, admin.Password, admin.MaxRoom, admin.MaxVoters, admin.IsActive, admin.CreatedAt, admin.UpdatedAt)
	return err
}

func (r *adminRepository) GetByID(id string) (*domain.Admin, error) {
	query := `
		SELECT id, username, password, max_room, max_voters, is_active, created_at, updated_at
		FROM admins
		WHERE id = $1
	`
	admin := &domain.Admin{}
	err := r.db.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Password,
		&admin.MaxRoom,
		&admin.MaxVoters,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *adminRepository) GetByUsername(username string) (*domain.Admin, error) {
	query := `
		SELECT id, username, password, max_room, max_voters, is_active, created_at, updated_at
		FROM admins
		WHERE username = $1
	`
	admin := &domain.Admin{}
	err := r.db.QueryRow(query, username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Password,
		&admin.MaxRoom,
		&admin.MaxVoters,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *adminRepository) Update(admin *domain.Admin) error {
	query := `
		UPDATE admins
		SET username = $2, password = $3, max_room = $4, max_voters = $5, is_active = $6, updated_at = $7
		WHERE id = $1
	`
	admin.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, admin.ID, admin.Username, admin.Password, admin.MaxRoom, admin.MaxVoters, admin.IsActive, admin.UpdatedAt)
	return err
}

func (r *adminRepository) Delete(id string) error {
	query := `DELETE FROM admins WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *adminRepository) GetRoomCount(adminID string) (int, error) {
	query := `SELECT COUNT(*) FROM rooms WHERE admin_id = $1`
	var count int
	err := r.db.QueryRow(query, adminID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *adminRepository) GetTotalVotersCount(adminID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(CASE 
			WHEN r.voters_type = 'custom_tickets' THEN (SELECT COUNT(*) FROM tickets WHERE room_id = r.id)
			WHEN r.voters_type = 'wild_limited' THEN COALESCE(r.voters_limit, 0)
			ELSE 0
		END), 0)
		FROM rooms r
		WHERE r.admin_id = $1
	`
	var count int
	err := r.db.QueryRow(query, adminID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

type loginAttemptRepository struct {
	db *sql.DB
}

func NewLoginAttemptRepository(db *sql.DB) domain.LoginAttemptRepository {
	return &loginAttemptRepository{db: db}
}

func (r *loginAttemptRepository) RecordAttempt(attempt *domain.LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (id, identifier, attempt_at, success)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query, attempt.ID, attempt.Identifier, attempt.AttemptAt, attempt.Success)
	return err
}

func (r *loginAttemptRepository) GetRecentFailedAttempts(identifier string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM login_attempts 
		WHERE identifier = $1 AND success = false AND attempt_at >= $2
	`
	var count int
	err := r.db.QueryRow(query, identifier, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *loginAttemptRepository) GetLastAttempt(identifier string) (*domain.LoginAttempt, error) {
	query := `
		SELECT id, identifier, attempt_at, success
		FROM login_attempts
		WHERE identifier = $1
		ORDER BY attempt_at DESC
		LIMIT 1
	`
	attempt := &domain.LoginAttempt{}
	err := r.db.QueryRow(query, identifier).Scan(
		&attempt.ID,
		&attempt.Identifier,
		&attempt.AttemptAt,
		&attempt.Success,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return attempt, nil
}

func (r *loginAttemptRepository) CleanupOldAttempts(before time.Time) error {
	query := `DELETE FROM login_attempts WHERE attempt_at < $1`
	_, err := r.db.Exec(query, before)
	return err
}
