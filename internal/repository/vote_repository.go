package repository

import (
	"database/sql"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
)

type voteRepository struct {
	db *sql.DB
}

func NewVoteRepository(db *sql.DB) domain.VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) Create(vote *domain.Vote) error {
	query := `
		INSERT INTO votes (id, room_id, candidate_id, sub_candidate_id, voter_identifier, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, vote.ID, vote.RoomID, vote.CandidateID, vote.SubCandidateID, vote.VoterIdentifier, vote.CreatedAt)
	return err
}

func (r *voteRepository) GetByRoomID(roomID string) ([]*domain.Vote, error) {
	query := `
		SELECT id, room_id, candidate_id, sub_candidate_id, voter_identifier, created_at
		FROM votes
		WHERE room_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []*domain.Vote
	for rows.Next() {
		vote := &domain.Vote{}
		err := rows.Scan(
			&vote.ID,
			&vote.RoomID,
			&vote.CandidateID,
			&vote.SubCandidateID,
			&vote.VoterIdentifier,
			&vote.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *voteRepository) GetVoteCountByRoom(roomID string) ([]*domain.VoteCount, error) {
	query := `
		SELECT candidate_id, COUNT(*) as count, MAX(created_at) as timestamp
		FROM votes
		WHERE room_id = $1
		GROUP BY candidate_id
		ORDER BY count DESC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voteCounts []*domain.VoteCount
	for rows.Next() {
		vc := &domain.VoteCount{}
		err := rows.Scan(&vc.CandidateID, &vc.Count, &vc.Timestamp)
		if err != nil {
			return nil, err
		}
		voteCounts = append(voteCounts, vc)
	}

	return voteCounts, nil
}

func (r *voteRepository) GetTotalVoteCountByRoom(roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM votes WHERE room_id = $1`
	var count int
	err := r.db.QueryRow(query, roomID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *voteRepository) CheckVoterHasVoted(roomID string, voterIdentifier string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM votes WHERE room_id = $1 AND voter_identifier = $2)`
	var exists bool
	err := r.db.QueryRow(query, roomID, voterIdentifier).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *voteRepository) GetRealtimeVoteCounts(roomID string) ([]*domain.VoteCount, error) {
	query := `
		SELECT candidate_id, COUNT(*) as count, NOW() as timestamp
		FROM votes
		WHERE room_id = $1
		GROUP BY candidate_id
		ORDER BY candidate_id
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voteCounts []*domain.VoteCount
	for rows.Next() {
		vc := &domain.VoteCount{}
		var timestamp time.Time
		err := rows.Scan(&vc.CandidateID, &vc.Count, &timestamp)
		if err != nil {
			return nil, err
		}
		vc.Timestamp = timestamp
		voteCounts = append(voteCounts, vc)
	}

	return voteCounts, nil
}
