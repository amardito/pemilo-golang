package http

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/amardito/pemilo-golang/internal/config"
	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const voteSessionTTL = 10 * time.Minute

const (
	defaultEventMaxVoters     = 1500
	defaultEventMaxCandidates = 12
)

type Server struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	cfg    config.Config
}

type voteClaims struct {
	EventID uuid.UUID `json:"event_id"`
	VoterID uuid.UUID `json:"voter_id"`
	Token   string    `json:"token"`
	jwt.RegisteredClaims
}

type contextKey string

const voteClaimsContextKey contextKey = "vote_claims"

func NewServer(logger *slog.Logger, db *pgxpool.Pool, cfg config.Config) *Server {
	return &Server{logger: logger, db: db, cfg: cfg}
}

func (s *Server) Router() nethttp.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/events", s.handleCreateEvent)
		r.Post("/events/{eventId}/slates", s.handleCreateSlate)
		r.Post("/events/{eventId}/voters/import", s.handleImportVoters)
		r.Post("/events/{eventId}/voters/tokens/generate", s.handleGenerateTokens)
		r.Get("/events/{eventId}/voters/tokens/export", s.handleExportTokens)
		r.Post("/events/{eventId}/vote/login", s.handleVoteLogin)
		r.With(s.voteSessionAuth).Post("/events/{eventId}/vote/submit", s.handleVoteSubmit)
		r.Get("/events/{eventId}/stats", s.handleStats)
	})

	return r
}

func (s *Server) handleCreateEvent(w nethttp.ResponseWriter, r *nethttp.Request) {
	type request struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		OpensAt     *time.Time `json:"opens_at"`
		ClosesAt    *time.Time `json:"closes_at"`
	}
	var req request
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "title is required")
		return
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "DRAFT"
	}
	switch status {
	case "DRAFT", "SCHEDULED", "OPEN", "CLOSED", "LOCKED":
	default:
		writeError(w, nethttp.StatusBadRequest, "validation_error", "invalid status")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	var eventID uuid.UUID
	err := s.db.QueryRow(ctx, `
		INSERT INTO events (title, description, status, opens_at, closes_at, ballot_type, max_voters, max_candidates)
		VALUES ($1, $2, $3, $4, $5, 'SECRET', $6, $7)
		RETURNING id`, req.Title, req.Description, status, req.OpensAt, req.ClosesAt, defaultEventMaxVoters, defaultEventMaxCandidates).Scan(&eventID)
	if err != nil {
		s.logger.Error("create event failed", "error", err)
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to create event")
		return
	}

	writeJSON(w, nethttp.StatusCreated, map[string]any{"id": eventID, "title": req.Title, "status": status})
}

func (s *Server) handleCreateSlate(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	type member struct {
		Role      string `json:"role"`
		FullName  string `json:"full_name"`
		PhotoURL  string `json:"photo_url"`
		Bio       string `json:"bio"`
		SortOrder int    `json:"sort_order"`
	}
	type request struct {
		Number  int      `json:"number"`
		Name    string   `json:"name"`
		Vision  string   `json:"vision"`
		Mission string   `json:"mission"`
		Members []member `json:"members"`
	}
	var req request
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "name is required")
		return
	}
	if req.Number <= 0 {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "number must be greater than 0")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	var slateID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO slates (event_id, number, name, vision, mission)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, eventID, req.Number, req.Name, req.Vision, req.Mission).Scan(&slateID)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "failed to create slate")
		return
	}

	for i, m := range req.Members {
		role := strings.TrimSpace(m.Role)
		fullName := strings.TrimSpace(m.FullName)
		if role == "" || fullName == "" {
			writeError(w, nethttp.StatusBadRequest, "validation_error", "member role and full_name are required")
			return
		}
		sortOrder := m.SortOrder
		if sortOrder == 0 {
			sortOrder = i + 1
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO slate_members (slate_id, role, full_name, photo_url, bio, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)`, slateID, role, fullName, strings.TrimSpace(m.PhotoURL), m.Bio, sortOrder)
		if err != nil {
			writeError(w, nethttp.StatusBadRequest, "validation_error", "failed to create slate member")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to commit transaction")
		return
	}

	writeJSON(w, nethttp.StatusCreated, map[string]any{"id": slateID, "event_id": eventID, "name": req.Name, "number": req.Number})
}

func (s *Server) handleImportVoters(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(16 << 20); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid_request", "failed to parse multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "file is required")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "failed to read csv header")
		return
	}

	headerMap := map[string]int{}
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}
	fullNameIdx, hasFullName := headerMap["full_name"]
	nimIdx, hasNIM := headerMap["nim"]
	classIdx, hasClass := headerMap["class_name"]
	if !hasFullName || !hasNIM {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "csv requires full_name and nim headers")
		return
	}

	type voter struct {
		fullName  string
		nim       string
		className string
	}
	voters := make([]voter, 0)
	seenNIM := map[string]int{}
	rowNum := 1

	for {
		rowNum++
		rec, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("csv parse error at row %d", rowNum))
			return
		}
		if fullNameIdx >= len(rec) || nimIdx >= len(rec) {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("missing columns at row %d", rowNum))
			return
		}

		fullName := strings.TrimSpace(rec[fullNameIdx])
		nim := domain.NormalizeNIM(rec[nimIdx])
		className := ""
		if hasClass && classIdx < len(rec) {
			className = strings.TrimSpace(rec[classIdx])
		}
		if fullName == "" {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("full_name required at row %d", rowNum))
			return
		}
		if nim == "" {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("nim required at row %d", rowNum))
			return
		}
		if len(nim) > 50 {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("nim max 50 chars at row %d", rowNum))
			return
		}
		if prevRow, exists := seenNIM[nim]; exists {
			writeError(w, nethttp.StatusBadRequest, "validation_error", fmt.Sprintf("duplicate nim %s in csv row %d and %d", nim, prevRow, rowNum))
			return
		}
		seenNIM[nim] = rowNum
		voters = append(voters, voter{fullName: fullName, nim: nim, className: className})
	}

	if len(voters) == 0 {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "csv has no voter rows")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	inserted := 0
	skipped := 0
	for _, v := range voters {
		ct, err := tx.Exec(ctx, `
			INSERT INTO voters (event_id, full_name, nim, class_name)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (event_id, nim) DO NOTHING`, eventID, v.fullName, v.nim, v.className)
		if err != nil {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to import voters")
			return
		}
		if ct.RowsAffected() == 1 {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to commit import")
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"event_id":       eventID,
		"imported_count": inserted,
		"skipped_count":  skipped,
	})
}

func (s *Server) handleGenerateTokens(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT v.id
		FROM voters v
		LEFT JOIN voter_tokens vt ON vt.voter_id = v.id
		WHERE v.event_id = $1 AND vt.id IS NULL
		ORDER BY v.created_at ASC`, eventID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to query voters")
		return
	}
	defer rows.Close()

	generated := 0
	for rows.Next() {
		var voterID uuid.UUID
		if err := rows.Scan(&voterID); err != nil {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to scan voter")
			return
		}

		persisted := false
		for attempts := 0; attempts < 5; attempts++ {
			token, err := generateToken(8)
			if err != nil {
				writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to generate token")
				return
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO voter_tokens (event_id, voter_id, token, status)
				VALUES ($1, $2, $3, 'ACTIVE')`, eventID, voterID, token)
			if err == nil {
				generated++
				persisted = true
				break
			}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to persist token")
			return
		}
		if !persisted {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to persist unique token")
			return
		}
	}
	if rows.Err() != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to iterate voters")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to commit token generation")
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{"event_id": eventID, "generated_count": generated})
}

func (s *Server) handleExportTokens(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT v.full_name, v.nim, COALESCE(v.class_name, ''), vt.token
		FROM voters v
		JOIN voter_tokens vt ON vt.voter_id = v.id
		WHERE v.event_id = $1
		ORDER BY v.full_name ASC`, eventID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to export tokens")
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"event-%s-tokens.csv\"", eventID.String()))
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	_ = csvWriter.Write([]string{"full_name", "nim", "class_name", "token"})
	for rows.Next() {
		var fullName, nim, className, token string
		if err := rows.Scan(&fullName, &nim, &className, &token); err != nil {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to scan token row")
			return
		}
		_ = csvWriter.Write([]string{fullName, nim, className, token})
	}
}

func (s *Server) handleVoteLogin(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	type request struct {
		Token string `json:"token"`
		NIM   string `json:"nim"`
	}
	var req request
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	token := strings.ToUpper(strings.TrimSpace(req.Token))
	nim := domain.NormalizeNIM(req.NIM)
	if err := domain.ValidateToken(token); err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if nim == "" || len(nim) > 50 {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "nim is required and max 50 chars")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	var eventTitle, eventStatus string
	var voterID uuid.UUID
	var hasVoted bool
	var tokenStatus string
	err := s.db.QueryRow(ctx, `
		SELECT e.title, e.status, v.id, v.has_voted, vt.status
		FROM voter_tokens vt
		JOIN voters v ON v.id = vt.voter_id AND v.event_id = vt.event_id
		JOIN events e ON e.id = v.event_id
		WHERE vt.event_id = $1 AND vt.token = $2 AND v.nim = $3`, eventID, token, nim).Scan(
		&eventTitle,
		&eventStatus,
		&voterID,
		&hasVoted,
		&tokenStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, nethttp.StatusUnauthorized, "invalid_credentials", "token or nim is invalid")
		return
	}
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to validate credentials")
		return
	}
	if eventStatus != "OPEN" {
		writeError(w, nethttp.StatusForbidden, "event_closed", "event is not open")
		return
	}
	if hasVoted || tokenStatus != "ACTIVE" {
		writeError(w, nethttp.StatusConflict, "already_voted", "vote already submitted")
		return
	}

	slates, err := s.getSlatesWithMembers(ctx, eventID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to get slates")
		return
	}

	session, err := s.signVoteSession(eventID, voterID, token)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to create session")
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"event": map[string]any{
			"id":     eventID,
			"title":  eventTitle,
			"status": eventStatus,
		},
		"slates":       slates,
		"vote_session": session,
	})
}

func (s *Server) handleVoteSubmit(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	claims, ok := r.Context().Value(voteClaimsContextKey).(*voteClaims)
	if !ok {
		writeError(w, nethttp.StatusUnauthorized, "unauthorized", "missing vote session")
		return
	}
	if claims.EventID != eventID {
		writeError(w, nethttp.StatusForbidden, "forbidden", "session does not match event")
		return
	}

	type request struct {
		SlateID string `json:"slate_id"`
	}
	var req request
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	slateID, err := uuid.Parse(req.SlateID)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "invalid slate_id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	var eventStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM events WHERE id = $1 FOR UPDATE`, eventID).Scan(&eventStatus); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to check event")
		return
	}
	if eventStatus != "OPEN" {
		writeError(w, nethttp.StatusForbidden, "event_closed", "event is not open")
		return
	}

	var hasVoted bool
	var tokenStatus string
	err = tx.QueryRow(ctx, `
		SELECT v.has_voted, vt.status
		FROM voters v
		JOIN voter_tokens vt ON vt.voter_id = v.id AND vt.event_id = v.event_id
		WHERE v.id = $1 AND v.event_id = $2 AND vt.token = $3
		FOR UPDATE OF v, vt`, claims.VoterID, eventID, claims.Token).Scan(&hasVoted, &tokenStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, nethttp.StatusUnauthorized, "invalid_session", "vote session is invalid")
		return
	}
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to lock voter")
		return
	}
	if hasVoted || tokenStatus != "ACTIVE" {
		writeError(w, nethttp.StatusConflict, "already_voted", "vote already submitted")
		return
	}

	var slateExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM slates WHERE id = $1 AND event_id = $2)`, slateID, eventID).Scan(&slateExists); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to validate slate")
		return
	}
	if !slateExists {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "slate not found")
		return
	}

	if _, err := tx.Exec(ctx, `INSERT INTO ballots (event_id, slate_id) VALUES ($1, $2)`, eventID, slateID); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to insert ballot")
		return
	}

	result, err := tx.Exec(ctx, `UPDATE voters SET has_voted = TRUE, voted_at = NOW() WHERE id = $1 AND has_voted = FALSE`, claims.VoterID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to update voter")
		return
	}
	if result.RowsAffected() != 1 {
		writeError(w, nethttp.StatusConflict, "already_voted", "vote already submitted")
		return
	}

	result, err = tx.Exec(ctx, `UPDATE voter_tokens SET status = 'USED', used_at = NOW() WHERE event_id = $1 AND voter_id = $2 AND token = $3 AND status = 'ACTIVE'`, eventID, claims.VoterID, claims.Token)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to update token")
		return
	}
	if result.RowsAffected() != 1 {
		writeError(w, nethttp.StatusConflict, "already_voted", "vote already submitted")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to commit vote")
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]string{"status": "vote_submitted"})
}

func (s *Server) handleStats(w nethttp.ResponseWriter, r *nethttp.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	var totalVoters, votedCount int
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) AS total_voters,
		       COALESCE(SUM(CASE WHEN has_voted THEN 1 ELSE 0 END), 0) AS voted_count
		FROM voters WHERE event_id = $1`, eventID).Scan(&totalVoters, &votedCount); err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to query voter stats")
		return
	}

	type voteBySlate struct {
		SlateID uuid.UUID `json:"slate_id"`
		Number  int       `json:"number"`
		Name    string    `json:"name"`
		Votes   int       `json:"votes"`
	}
	votes := make([]voteBySlate, 0)
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.number, s.name, COALESCE(COUNT(b.id), 0) AS votes
		FROM slates s
		LEFT JOIN ballots b ON b.slate_id = s.id AND b.event_id = s.event_id
		WHERE s.event_id = $1
		GROUP BY s.id, s.number, s.name
		ORDER BY s.number ASC`, eventID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to query votes")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item voteBySlate
		if err := rows.Scan(&item.SlateID, &item.Number, &item.Name, &item.Votes); err != nil {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to read vote stats")
			return
		}
		votes = append(votes, item)
	}

	type latestVoter struct {
		FullName string     `json:"full_name"`
		NIM      string     `json:"nim"`
		VotedAt  *time.Time `json:"voted_at"`
	}
	latest := make([]latestVoter, 0)
	rows, err = s.db.Query(ctx, `
		SELECT full_name, nim, voted_at
		FROM voters
		WHERE event_id = $1 AND has_voted = TRUE
		ORDER BY voted_at DESC
		LIMIT 10`, eventID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to query latest voters")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item latestVoter
		if err := rows.Scan(&item.FullName, &item.NIM, &item.VotedAt); err != nil {
			writeError(w, nethttp.StatusInternalServerError, "internal_error", "failed to read latest voters")
			return
		}
		latest = append(latest, item)
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"event_id":        eventID,
		"total_voters":    totalVoters,
		"voted_count":     votedCount,
		"not_voted_count": totalVoters - votedCount,
		"votes_by_slate":  votes,
		"latest_voters":   latest,
	})
}

func (s *Server) voteSessionAuth(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			writeError(w, nethttp.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		if len(header) <= 7 {
			writeError(w, nethttp.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		raw := strings.TrimSpace(header[7:])
		if raw == "" {
			writeError(w, nethttp.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		token, err := jwt.ParseWithClaims(raw, &voteClaims{}, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid signing method")
			}
			return []byte(s.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, nethttp.StatusUnauthorized, "unauthorized", "invalid vote session")
			return
		}
		claims, ok := token.Claims.(*voteClaims)
		if !ok {
			writeError(w, nethttp.StatusUnauthorized, "unauthorized", "invalid vote session claims")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), voteClaimsContextKey, claims)))
	})
}

func (s *Server) signVoteSession(eventID, voterID uuid.UUID, token string) (string, error) {
	now := time.Now()
	claims := voteClaims{
		EventID: eventID,
		VoterID: voterID,
		Token:   token,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "pemilo-backend",
			Subject:   voterID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(voteSessionTTL)),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Server) getSlatesWithMembers(ctx context.Context, eventID uuid.UUID) ([]map[string]any, error) {
	slateRows, err := s.db.Query(ctx, `
		SELECT id, number, name, COALESCE(vision, ''), COALESCE(mission, '')
		FROM slates WHERE event_id = $1 ORDER BY number ASC`, eventID)
	if err != nil {
		return nil, err
	}
	defer slateRows.Close()

	slates := make([]map[string]any, 0)
	for slateRows.Next() {
		var slateID uuid.UUID
		var number int
		var name, vision, mission string
		if err := slateRows.Scan(&slateID, &number, &name, &vision, &mission); err != nil {
			return nil, err
		}

		memberRows, err := s.db.Query(ctx, `
			SELECT role, full_name, COALESCE(photo_url, ''), COALESCE(bio, ''), sort_order
			FROM slate_members WHERE slate_id = $1 ORDER BY sort_order ASC`, slateID)
		if err != nil {
			return nil, err
		}
		members := make([]map[string]any, 0)
		for memberRows.Next() {
			var role, fullName, photoURL, bio string
			var sortOrder int
			if err := memberRows.Scan(&role, &fullName, &photoURL, &bio, &sortOrder); err != nil {
				memberRows.Close()
				return nil, err
			}
			members = append(members, map[string]any{
				"role":       role,
				"full_name":  fullName,
				"photo_url":  photoURL,
				"bio":        bio,
				"sort_order": sortOrder,
			})
		}
		memberRows.Close()

		slates = append(slates, map[string]any{
			"id":      slateID,
			"number":  number,
			"name":    name,
			"vision":  vision,
			"mission": mission,
			"members": members,
		})
	}

	return slates, nil
}

func parseEventID(w nethttp.ResponseWriter, r *nethttp.Request) (uuid.UUID, bool) {
	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "validation_error", "invalid eventId")
		return uuid.Nil, false
	}
	return eventID, true
}

func writeJSON(w nethttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w nethttp.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func decodeJSON(r *nethttp.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func generateToken(length int) (string, error) {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String(), nil
}
