package domain

import "time"

// VotersType defines the type of voter validation for a room
type VotersType string

const (
	VotersTypeCustomTickets VotersType = "custom_tickets"
	VotersTypeWildLimited   VotersType = "wild_limited"
	VotersTypeWildUnlimited VotersType = "wild_unlimited"
)

// RoomStatus defines whether a room is active
type RoomStatus string

const (
	RoomStatusEnabled  RoomStatus = "enabled"
	RoomStatusDisabled RoomStatus = "disabled"
)

// PublishState defines the publication state of a room
type PublishState string

const (
	PublishStateDraft     PublishState = "draft"
	PublishStatePublished PublishState = "published"
)

// SessionState defines the current state of a voting session
type SessionState string

const (
	SessionStateOpen   SessionState = "open"
	SessionStateClosed SessionState = "closed"
)

// Room represents an election/voting room
type Room struct {
	ID               string       `json:"id"`
	AdminID          string       `json:"admin_id"` // Owner of this room
	Name             string       `json:"name"`
	VotersType       VotersType   `json:"voters_type"`
	VotersLimit      *int         `json:"voters_limit,omitempty"`
	SessionStartTime *time.Time   `json:"session_start_time,omitempty"`
	SessionEndTime   *time.Time   `json:"session_end_time,omitempty"`
	Status           RoomStatus   `json:"status"`
	PublishState     PublishState `json:"publish_state"`
	SessionState     SessionState `json:"session_state"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// Validate performs validation on Room based on voters_type
func (r *Room) Validate() error {
	if r.Name == "" {
		return ErrInvalidRoomName
	}

	switch r.VotersType {
	case VotersTypeWildLimited:
		if r.VotersLimit == nil || *r.VotersLimit <= 0 {
			return ErrVotersLimitRequired
		}
	case VotersTypeWildUnlimited:
		if r.SessionStartTime == nil || r.SessionEndTime == nil {
			return ErrSessionRangeRequired
		}
		if r.SessionEndTime.Before(*r.SessionStartTime) {
			return ErrInvalidSessionRange
		}
	case VotersTypeCustomTickets:
		// No additional validation for custom tickets
	default:
		return ErrInvalidVotersType
	}

	return nil
}

// IsSessionActive checks if the current time is within session active range
func (r *Room) IsSessionActive() bool {
	if r.VotersType != VotersTypeWildUnlimited {
		return true
	}

	now := time.Now()
	return r.SessionStartTime != nil && r.SessionEndTime != nil &&
		!now.Before(*r.SessionStartTime) && !now.After(*r.SessionEndTime)
}

// RoomRepository defines the interface for room persistence
type RoomRepository interface {
	Create(room *Room) error
	GetByID(id string) (*Room, error)
	Update(room *Room) error
	Delete(id string) error
	List(filters RoomFilters) ([]*Room, error)
	UpdateSessionState(roomID string, state SessionState) error
	CountByAdminID(adminID string) (int, error)
}

// RoomFilters defines filters for listing rooms
type RoomFilters struct {
	Status       *RoomStatus
	PublishState *PublishState
	SessionState *SessionState
}
