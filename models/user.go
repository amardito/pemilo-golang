package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents admin users who manage voting rooms
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string         `json:"-" gorm:"not null"` // Never expose in JSON
	CreatedAt    time.Time      `json:"created_at"`
	LastLogin    *time.Time     `json:"last_login"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	VotingRooms []VotingRoom `json:"voting_rooms,omitempty" gorm:"foreignKey:CreatedBy"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the provided password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Session stores server-side sessions for users
type Session struct {
	ID        string         `json:"id" gorm:"primaryKey;size:64"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Data      string         `json:"-" gorm:"type:text"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
