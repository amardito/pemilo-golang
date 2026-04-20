package dto

import "time"

// ── Auth ──

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ── Event ──

type CreateEventRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	OpensAt     *time.Time `json:"opens_at"`
	ClosesAt    *time.Time `json:"closes_at"`
}

type UpdateEventRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	OpensAt     *time.Time `json:"opens_at"`
	ClosesAt    *time.Time `json:"closes_at"`
}

// ── Slate ──

type CreateSlateRequest struct {
	Number   int     `json:"number" binding:"required,min=1"`
	Name     string  `json:"name" binding:"required"`
	Vision   *string `json:"vision"`
	Mission  *string `json:"mission"`
	PhotoURL *string `json:"photo_url"`
}

type UpdateSlateRequest struct {
	Number   *int    `json:"number"`
	Name     *string `json:"name"`
	Vision   *string `json:"vision"`
	Mission  *string `json:"mission"`
	PhotoURL *string `json:"photo_url"`
}

// ── Slate Member ──

type CreateSlateMemberRequest struct {
	Role      string  `json:"role" binding:"required"`
	FullName  string  `json:"full_name" binding:"required"`
	PhotoURL  *string `json:"photo_url"`
	Bio       *string `json:"bio"`
	SortOrder *int    `json:"sort_order"`
}

type UpdateSlateMemberRequest struct {
	Role      *string `json:"role"`
	FullName  *string `json:"full_name"`
	PhotoURL  *string `json:"photo_url"`
	Bio       *string `json:"bio"`
	SortOrder *int    `json:"sort_order"`
}

// ── Voting (Public) ──

type VotePrepareRequest struct {
	Token string `json:"token" binding:"required"`
	NIM   string `json:"nim" binding:"required"`
}

type VotePrepareResponse struct {
	OK           bool          `json:"ok"`
	VoterDisplay VoterDisplay  `json:"voter_display"`
	Slates       []SlatePublic `json:"slates"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

type VoterDisplay struct {
	FullName  string  `json:"full_name"`
	ClassName *string `json:"class_name"`
}

type SlatePublic struct {
	ID       string              `json:"id"`
	Number   int                 `json:"number"`
	Name     string              `json:"name"`
	Vision   *string             `json:"vision"`
	Mission  *string             `json:"mission"`
	PhotoURL *string             `json:"photo_url"`
	Members  []SlateMemberPublic `json:"members"`
}

type SlateMemberPublic struct {
	Role      string  `json:"role"`
	FullName  string  `json:"full_name"`
	PhotoURL  *string `json:"photo_url"`
	Bio       *string `json:"bio"`
	SortOrder int     `json:"sort_order"`
}

type VoteSubmitRequest struct {
	Token   string `json:"token" binding:"required"`
	NIM     string `json:"nim" binding:"required"`
	SlateID string `json:"slate_id" binding:"required,uuid"`
}

// ── Stats ──

type StatsResponse struct {
	EventID       string        `json:"event_id"`
	TotalVoters   int           `json:"total_voters"`
	VotedCount    int           `json:"voted_count"`
	NotVotedCount int           `json:"not_voted_count"`
	VotesBySlate  []SlateVotes  `json:"votes_by_slate"`
	LatestVoters  []LatestVoter `json:"latest_voters"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type SlateVotes struct {
	SlateID string `json:"slate_id"`
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Votes   int    `json:"votes"`
}

type LatestVoter struct {
	FullName  string    `json:"full_name"`
	ClassName *string   `json:"class_name"`
	VotedAt   time.Time `json:"voted_at"`
}

// ── Voter Import ──

type ImportResult struct {
	ImportedCount int            `json:"imported_count"`
	Rejected      []ImportReject `json:"rejected"`
}

type ImportReject struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// ── Voter List ──

type VoterListParams struct {
	Query    string
	Status   string
	HasVoted *bool
	Page     int
	PerPage  int
}

type VoterListResponse struct {
	Voters  []VoterDTO `json:"voters"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
}

type VoterDTO struct {
	ID        string     `json:"id"`
	FullName  string     `json:"full_name"`
	NIMRaw    string     `json:"nim_raw"`
	ClassName *string    `json:"class_name"`
	HasVoted  bool       `json:"has_voted"`
	VotedAt   *time.Time `json:"voted_at"`
	Status    string     `json:"status"`
}

// ── Payment ──

type UpgradeRequest struct {
	Package string `json:"package" binding:"required,oneof=STARTER PRO"`
}

type UpgradeResponse struct {
	PaymentURL string `json:"payment_url"`
	OrderID    string `json:"order_id"`
}

type OrderDTO struct {
	ID              string    `json:"id"`
	EventID         string    `json:"event_id"`
	Package         string    `json:"package"`
	Amount          int       `json:"amount"`
	Status          string    `json:"status"`
	IPaymuReference *string   `json:"ipaymu_reference"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ── Generic ──

type SuccessResponse struct {
	OK      bool        `json:"ok"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// ── Audit Log ──

type AuditLogDTO struct {
	ID          string    `json:"id"`
	Action      string    `json:"action"`
	ActorUserID *string   `json:"actor_user_id"`
	Meta        string    `json:"meta"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLogListResponse struct {
	Logs    []AuditLogDTO `json:"logs"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}
