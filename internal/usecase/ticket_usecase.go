package usecase

import (
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/google/uuid"
)

type TicketUsecase struct {
	ticketRepo domain.TicketRepository
	roomRepo   domain.RoomRepository
}

func NewTicketUsecase(ticketRepo domain.TicketRepository, roomRepo domain.RoomRepository) *TicketUsecase {
	return &TicketUsecase{
		ticketRepo: ticketRepo,
		roomRepo:   roomRepo,
	}
}

func (u *TicketUsecase) CreateTicket(roomID, code string) (*domain.Ticket, error) {
	// Verify room exists and is custom_tickets type
	room, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, err
	}

	if room.VotersType != domain.VotersTypeCustomTickets {
		return nil, domain.ErrInvalidVotersType
	}

	ticket := &domain.Ticket{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		Code:      code,
		IsUsed:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.ticketRepo.Create(ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (u *TicketUsecase) CreateTicketsBulk(roomID string, codes []string) ([]*domain.Ticket, error) {
	// Verify room exists and is custom_tickets type
	room, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, err
	}

	if room.VotersType != domain.VotersTypeCustomTickets {
		return nil, domain.ErrInvalidVotersType
	}

	tickets := make([]*domain.Ticket, 0, len(codes))
	for _, code := range codes {
		ticket := &domain.Ticket{
			ID:        uuid.New().String(),
			RoomID:    roomID,
			Code:      code,
			IsUsed:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tickets = append(tickets, ticket)
	}

	if err := u.ticketRepo.CreateBulk(tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (u *TicketUsecase) GetTicketsByRoom(roomID string) ([]*domain.Ticket, error) {
	return u.ticketRepo.GetByRoomID(roomID)
}

func (u *TicketUsecase) VerifyTicket(roomID, code string) (*domain.Ticket, error) {
	ticket, err := u.ticketRepo.GetByCode(roomID, code)
	if err != nil {
		return nil, err
	}

	if ticket.IsUsed {
		return nil, domain.ErrTicketAlreadyUsed
	}

	return ticket, nil
}

func (u *TicketUsecase) DeleteTicket(id string) error {
	return u.ticketRepo.Delete(id)
}
