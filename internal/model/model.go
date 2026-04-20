package model

import "time"

// ── Enums ──

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusScheduled EventStatus = "SCHEDULED"
	EventStatusOpen      EventStatus = "OPEN"
	EventStatusClosed    EventStatus = "CLOSED"
	EventStatusLocked    EventStatus = "LOCKED"
)

type VoterStatus string

const (
	VoterStatusEligible VoterStatus = "ELIGIBLE"
	VoterStatusDisabled VoterStatus = "DISABLED"
)

type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "ACTIVE"
	TokenStatusUsed    TokenStatus = "USED"
	TokenStatusRevoked TokenStatus = "REVOKED"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "PENDING"
	OrderStatusPaid    OrderStatus = "PAID"
	OrderStatusExpired OrderStatus = "EXPIRED"
	OrderStatusFailed  OrderStatus = "FAILED"
)

type Package string

const (
	PackageFree    Package = "FREE"
	PackageStarter Package = "STARTER"
	PackagePro     Package = "PRO"
)

// ── Domain Models ──

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Event struct {
	ID          string      `json:"id" db:"id"`
	OwnerUserID string      `json:"owner_user_id" db:"owner_user_id"`
	Title       string      `json:"title" db:"title"`
	Description *string     `json:"description" db:"description"`
	Status      EventStatus `json:"status" db:"status"`
	OpensAt     *time.Time  `json:"opens_at" db:"opens_at"`
	ClosesAt    *time.Time  `json:"closes_at" db:"closes_at"`
	MaxSlates   int         `json:"max_slates" db:"max_slates"`
	MaxVoters   int         `json:"max_voters" db:"max_voters"`
	Package     Package     `json:"package" db:"package"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type Slate struct {
	ID        string        `json:"id" db:"id"`
	EventID   string        `json:"event_id" db:"event_id"`
	Number    int           `json:"number" db:"number"`
	Name      string        `json:"name" db:"name"`
	Vision    *string       `json:"vision" db:"vision"`
	Mission   *string       `json:"mission" db:"mission"`
	PhotoURL  *string       `json:"photo_url" db:"photo_url"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	Members   []SlateMember `json:"members,omitempty"`
}

type SlateMember struct {
	ID        string  `json:"id" db:"id"`
	SlateID   string  `json:"slate_id" db:"slate_id"`
	Role      string  `json:"role" db:"role"`
	FullName  string  `json:"full_name" db:"full_name"`
	PhotoURL  *string `json:"photo_url" db:"photo_url"`
	Bio       *string `json:"bio" db:"bio"`
	SortOrder int     `json:"sort_order" db:"sort_order"`
}

type Voter struct {
	ID            string      `json:"id" db:"id"`
	EventID       string      `json:"event_id" db:"event_id"`
	FullName      string      `json:"full_name" db:"full_name"`
	NIMRaw        string      `json:"nim_raw" db:"nim_raw"`
	NIMNormalized string      `json:"nim_normalized" db:"nim_normalized"`
	ClassName     *string     `json:"class_name" db:"class_name"`
	Status        VoterStatus `json:"status" db:"status"`
	HasVoted      bool        `json:"has_voted" db:"has_voted"`
	VotedAt       *time.Time  `json:"voted_at" db:"voted_at"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

type VoterToken struct {
	ID       string      `json:"id" db:"id"`
	EventID  string      `json:"event_id" db:"event_id"`
	VoterID  string      `json:"voter_id" db:"voter_id"`
	Token    string      `json:"token" db:"token"`
	Status   TokenStatus `json:"status" db:"status"`
	IssuedAt time.Time   `json:"issued_at" db:"issued_at"`
	UsedAt   *time.Time  `json:"used_at" db:"used_at"`
}

type Ballot struct {
	ID        string    `json:"id" db:"id"`
	EventID   string    `json:"event_id" db:"event_id"`
	SlateID   string    `json:"slate_id" db:"slate_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AuditLog struct {
	ID          string    `json:"id" db:"id"`
	EventID     string    `json:"event_id" db:"event_id"`
	ActorUserID *string   `json:"actor_user_id" db:"actor_user_id"`
	Action      string    `json:"action" db:"action"`
	Meta        string    `json:"meta" db:"meta"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Order struct {
	ID              string      `json:"id" db:"id"`
	EventID         string      `json:"event_id" db:"event_id"`
	Package         Package     `json:"package" db:"package"`
	Amount          int         `json:"amount" db:"amount"`
	Status          OrderStatus `json:"status" db:"status"`
	IPaymuReference *string     `json:"ipaymu_reference" db:"ipaymu_reference"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

// ── Package Limits ──

type PackageLimits struct {
	MaxSlates int
	MaxVoters int
	Price     int // rupiah, 0 for FREE
}

var PackageLimitsMap = map[Package]PackageLimits{
	PackageFree:    {MaxSlates: 2, MaxVoters: 30, Price: 0},
	PackageStarter: {MaxSlates: 6, MaxVoters: 200, Price: 79000},
	PackagePro:     {MaxSlates: 12, MaxVoters: 1500, Price: 149000},
}
