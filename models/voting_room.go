package models

import (
	"time"

	"gorm.io/gorm"
)

// RoomStatus represents the status of a voting room
type RoomStatus string

const (
	RoomStatusInactive RoomStatus = "inactive"
	RoomStatusActive   RoomStatus = "active"
	RoomStatusFinished RoomStatus = "finished"
)

// VotingRoom represents a voting session/event
type VotingRoom struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      RoomStatus     `json:"status" gorm:"size:20;default:'inactive';not null"`
	CreatedBy   uint           `json:"created_by" gorm:"not null;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Creator    User        `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Candidates []Candidate `json:"candidates,omitempty" gorm:"foreignKey:RoomID"`
	Tickets    []Ticket    `json:"tickets,omitempty" gorm:"foreignKey:RoomID"`
}

// IsActive returns true if the room is in active status
func (r *VotingRoom) IsActive() bool {
	return r.Status == RoomStatusActive
}

// CanVote returns true if voting is allowed in this room
func (r *VotingRoom) CanVote() bool {
	return r.Status == RoomStatusActive
}
