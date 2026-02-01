package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	TicketStatusUnused TicketStatus = "unused"
	TicketStatusUsed   TicketStatus = "used"
)

// Ticket represents a voting ticket/token
type Ticket struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"uniqueIndex;size:20;not null"`
	RoomID    uint           `json:"room_id" gorm:"not null;index"`
	VoterName *string        `json:"voter_name" gorm:"size:255"`
	Status    TicketStatus   `json:"status" gorm:"size:20;default:'unused';not null"`
	UsedAt    *time.Time     `json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Room VotingRoom `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Vote *Vote      `json:"vote,omitempty" gorm:"foreignKey:TicketID"`
}

// GenerateCode creates a secure random ticket code
func GenerateTicketCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsUsed returns true if the ticket has been used
func (t *Ticket) IsUsed() bool {
	return t.Status == TicketStatusUsed
}

// MarkUsed marks the ticket as used
func (t *Ticket) MarkUsed() {
	now := time.Now()
	t.Status = TicketStatusUsed
	t.UsedAt = &now
}
