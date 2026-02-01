package models

import (
	"time"

	"gorm.io/gorm"
)

// Vote represents a single vote cast using a ticket
type Vote struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TicketID    uint           `json:"ticket_id" gorm:"uniqueIndex;not null"` // One vote per ticket
	CandidateID uint           `json:"candidate_id" gorm:"not null;index"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Ticket    Ticket    `json:"ticket,omitempty" gorm:"foreignKey:TicketID"`
	Candidate Candidate `json:"candidate,omitempty" gorm:"foreignKey:CandidateID"`
}
