package domain

import "time"

// Ticket represents a voting ticket for custom_tickets rooms
type Ticket struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"room_id"`
	Code      string     `json:"code"`
	IsUsed    bool       `json:"is_used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TicketRepository defines the interface for ticket persistence
type TicketRepository interface {
	Create(ticket *Ticket) error
	CreateBulk(tickets []*Ticket) error
	GetByCode(roomID string, code string) (*Ticket, error)
	MarkAsUsed(id string) error
	GetByRoomID(roomID string) ([]*Ticket, error)
	Delete(id string) error
}
