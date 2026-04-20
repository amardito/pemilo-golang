package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/amard/pemilo-golang/internal/dto"
)

type BallotRepo struct {
	db *sql.DB
}

func NewBallotRepo(db *sql.DB) *BallotRepo {
	return &BallotRepo{db: db}
}

// InsertInTx inserts a ballot within a transaction — NO voter_id (secret ballot).
func (r *BallotRepo) InsertInTx(ctx context.Context, tx *sql.Tx, eventID, slateID string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO ballots (event_id, slate_id) VALUES ($1, $2)`,
		eventID, slateID,
	)
	return err
}

// GetVotesBySlate returns vote counts grouped by slate for an event.
func (r *BallotRepo) GetVotesBySlate(ctx context.Context, eventID string) ([]dto.SlateVotes, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.id, s.number, s.name, COUNT(b.id)
		 FROM slates s
		 LEFT JOIN ballots b ON b.slate_id = s.id AND b.event_id = s.event_id
		 WHERE s.event_id = $1
		 GROUP BY s.id, s.number, s.name
		 ORDER BY s.number`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.SlateVotes
	for rows.Next() {
		var sv dto.SlateVotes
		if err := rows.Scan(&sv.SlateID, &sv.Number, &sv.Name, &sv.Votes); err != nil {
			return nil, err
		}
		result = append(result, sv)
	}
	return result, rows.Err()
}

// GetTurnoutCounts returns total voters and voted count.
func (r *BallotRepo) GetTurnoutCounts(ctx context.Context, eventID string) (total int, voted int, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE has_voted = true) FROM voters WHERE event_id = $1`,
		eventID,
	).Scan(&total, &voted)
	return
}

// GetLatestVoters returns the most recent voters who have voted.
func (r *BallotRepo) GetLatestVoters(ctx context.Context, eventID string, limit int) ([]dto.LatestVoter, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT full_name, class_name, voted_at
		 FROM voters
		 WHERE event_id = $1 AND has_voted = true
		 ORDER BY voted_at DESC
		 LIMIT $2`,
		eventID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.LatestVoter
	for rows.Next() {
		var lv dto.LatestVoter
		var votedAt time.Time
		if err := rows.Scan(&lv.FullName, &lv.ClassName, &votedAt); err != nil {
			return nil, err
		}
		lv.VotedAt = votedAt
		result = append(result, lv)
	}
	return result, rows.Err()
}
