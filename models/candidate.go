package models

import (
	"time"

	"gorm.io/gorm"
)

// Candidate represents a candidate in a voting room
type Candidate struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	RoomID            uint           `json:"room_id" gorm:"not null;index"`
	ParentCandidateID *uint          `json:"parent_candidate_id" gorm:"index"` // For hierarchical candidates
	Name              string         `json:"name" gorm:"size:255;not null"`
	PhotoURL          string         `json:"photo_url" gorm:"size:500"`
	Description       string         `json:"description" gorm:"type:text"`
	DisplayOrder      int            `json:"display_order" gorm:"default:0"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Room     VotingRoom  `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Parent   *Candidate  `json:"parent,omitempty" gorm:"foreignKey:ParentCandidateID"`
	Children []Candidate `json:"children,omitempty" gorm:"foreignKey:ParentCandidateID"`
	Votes    []Vote      `json:"votes,omitempty" gorm:"foreignKey:CandidateID"`

	// Computed field (not stored in DB)
	VoteCount int `json:"vote_count" gorm:"-"`
}
