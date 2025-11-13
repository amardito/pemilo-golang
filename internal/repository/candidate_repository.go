package repository

import (
	"database/sql"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
)

type candidateRepository struct {
	db *sql.DB
}

func NewCandidateRepository(db *sql.DB) domain.CandidateRepository {
	return &candidateRepository{db: db}
}

func (r *candidateRepository) Create(candidate *domain.Candidate) error {
	query := `
		INSERT INTO candidates (id, room_id, name, photo_url, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, candidate.ID, candidate.RoomID, candidate.Name, candidate.PhotoURL, candidate.Description, candidate.CreatedAt, candidate.UpdatedAt)
	return err
}

func (r *candidateRepository) GetByID(id string) (*domain.Candidate, error) {
	query := `
		SELECT id, room_id, name, photo_url, description, created_at, updated_at
		FROM candidates
		WHERE id = $1
	`
	candidate := &domain.Candidate{}
	err := r.db.QueryRow(query, id).Scan(
		&candidate.ID,
		&candidate.RoomID,
		&candidate.Name,
		&candidate.PhotoURL,
		&candidate.Description,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCandidateNotFound
	}
	if err != nil {
		return nil, err
	}
	return candidate, nil
}

func (r *candidateRepository) GetByRoomID(roomID string) ([]*domain.Candidate, error) {
	query := `
		SELECT id, room_id, name, photo_url, description, created_at, updated_at
		FROM candidates
		WHERE room_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*domain.Candidate
	for rows.Next() {
		candidate := &domain.Candidate{}
		err := rows.Scan(
			&candidate.ID,
			&candidate.RoomID,
			&candidate.Name,
			&candidate.PhotoURL,
			&candidate.Description,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (r *candidateRepository) Update(candidate *domain.Candidate) error {
	query := `
		UPDATE candidates
		SET name = $2, photo_url = $3, description = $4, updated_at = $5
		WHERE id = $1
	`
	candidate.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, candidate.ID, candidate.Name, candidate.PhotoURL, candidate.Description, candidate.UpdatedAt)
	return err
}

func (r *candidateRepository) Delete(id string) error {
	query := `DELETE FROM candidates WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

type subCandidateRepository struct {
	db *sql.DB
}

func NewSubCandidateRepository(db *sql.DB) domain.SubCandidateRepository {
	return &subCandidateRepository{db: db}
}

func (r *subCandidateRepository) Create(subCandidate *domain.SubCandidate) error {
	query := `
		INSERT INTO sub_candidates (id, candidate_id, name, photo_url, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, subCandidate.ID, subCandidate.CandidateID, subCandidate.Name, subCandidate.PhotoURL, subCandidate.Description, subCandidate.CreatedAt, subCandidate.UpdatedAt)
	return err
}

func (r *subCandidateRepository) GetByID(id string) (*domain.SubCandidate, error) {
	query := `
		SELECT id, candidate_id, name, photo_url, description, created_at, updated_at
		FROM sub_candidates
		WHERE id = $1
	`
	subCandidate := &domain.SubCandidate{}
	err := r.db.QueryRow(query, id).Scan(
		&subCandidate.ID,
		&subCandidate.CandidateID,
		&subCandidate.Name,
		&subCandidate.PhotoURL,
		&subCandidate.Description,
		&subCandidate.CreatedAt,
		&subCandidate.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCandidateNotFound
	}
	if err != nil {
		return nil, err
	}
	return subCandidate, nil
}

func (r *subCandidateRepository) GetByCandidateID(candidateID string) ([]*domain.SubCandidate, error) {
	query := `
		SELECT id, candidate_id, name, photo_url, description, created_at, updated_at
		FROM sub_candidates
		WHERE candidate_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subCandidates []*domain.SubCandidate
	for rows.Next() {
		subCandidate := &domain.SubCandidate{}
		err := rows.Scan(
			&subCandidate.ID,
			&subCandidate.CandidateID,
			&subCandidate.Name,
			&subCandidate.PhotoURL,
			&subCandidate.Description,
			&subCandidate.CreatedAt,
			&subCandidate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subCandidates = append(subCandidates, subCandidate)
	}

	return subCandidates, nil
}

func (r *subCandidateRepository) Update(subCandidate *domain.SubCandidate) error {
	query := `
		UPDATE sub_candidates
		SET name = $2, photo_url = $3, description = $4, updated_at = $5
		WHERE id = $1
	`
	subCandidate.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, subCandidate.ID, subCandidate.Name, subCandidate.PhotoURL, subCandidate.Description, subCandidate.UpdatedAt)
	return err
}

func (r *subCandidateRepository) Delete(id string) error {
	query := `DELETE FROM sub_candidates WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
