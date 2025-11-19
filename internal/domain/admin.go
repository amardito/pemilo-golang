package domain

import "time"

// Admin represents an administrator account with quotas
type Admin struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Encrypted password, never expose in JSON
	MaxRoom   int       `json:"max_room"`
	MaxVoters int       `json:"max_voters"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginAttempt tracks failed login attempts for rate limiting
type LoginAttempt struct {
	ID         string    `json:"id"`
	Identifier string    `json:"identifier"` // Username or IP address
	AttemptAt  time.Time `json:"attempt_at"`
	Success    bool      `json:"success"`
}

// AdminRepository defines the interface for admin persistence
type AdminRepository interface {
	Create(admin *Admin) error
	GetByID(id string) (*Admin, error)
	GetByUsername(username string) (*Admin, error)
	Update(admin *Admin) error
	Delete(id string) error
	GetRoomCount(adminID string) (int, error)
	GetTotalVotersCount(adminID string) (int, error)
}

// LoginAttemptRepository defines the interface for login attempt tracking
type LoginAttemptRepository interface {
	RecordAttempt(attempt *LoginAttempt) error
	GetRecentFailedAttempts(identifier string, since time.Time) (int, error)
	GetLastAttempt(identifier string) (*LoginAttempt, error)
	CleanupOldAttempts(before time.Time) error
}
