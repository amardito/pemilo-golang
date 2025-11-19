package usecase

import (
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/google/uuid"
)

type RoomUsecase struct {
	roomRepo domain.RoomRepository
}

func NewRoomUsecase(roomRepo domain.RoomRepository) *RoomUsecase {
	return &RoomUsecase{
		roomRepo: roomRepo,
	}
}

func (u *RoomUsecase) CreateRoom(adminID, name string, votersType domain.VotersType, votersLimit *int, sessionStartTime, sessionEndTime *time.Time, status domain.RoomStatus, publishState domain.PublishState) (*domain.Room, error) {
	room := &domain.Room{
		ID:               uuid.New().String(),
		AdminID:          adminID,
		Name:             name,
		VotersType:       votersType,
		VotersLimit:      votersLimit,
		SessionStartTime: sessionStartTime,
		SessionEndTime:   sessionEndTime,
		Status:           status,
		PublishState:     publishState,
		SessionState:     domain.SessionStateOpen,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Validate room based on voters_type
	if err := room.Validate(); err != nil {
		return nil, err
	}

	if err := u.roomRepo.Create(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (u *RoomUsecase) GetRoom(id string) (*domain.Room, error) {
	return u.roomRepo.GetByID(id)
}

func (u *RoomUsecase) UpdateRoom(id string, name *string, votersType *domain.VotersType, votersLimit *int, sessionStartTime, sessionEndTime *time.Time, status *domain.RoomStatus, publishState *domain.PublishState) (*domain.Room, error) {
	room, err := u.roomRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name != nil {
		room.Name = *name
	}
	if votersType != nil {
		room.VotersType = *votersType
	}
	if votersLimit != nil {
		room.VotersLimit = votersLimit
	}
	if sessionStartTime != nil {
		room.SessionStartTime = sessionStartTime
	}
	if sessionEndTime != nil {
		room.SessionEndTime = sessionEndTime
	}
	if status != nil {
		room.Status = *status
	}
	if publishState != nil {
		room.PublishState = *publishState
	}

	// Validate updated room
	if err := room.Validate(); err != nil {
		return nil, err
	}

	if err := u.roomRepo.Update(room); err != nil {
		return nil, err
	}

	return room, nil
}

func (u *RoomUsecase) DeleteRoom(id string) error {
	return u.roomRepo.Delete(id)
}

func (u *RoomUsecase) ListRooms(filters domain.RoomFilters) ([]*domain.Room, error) {
	return u.roomRepo.List(filters)
}

func (u *RoomUsecase) CloseRoomSession(roomID string) error {
	return u.roomRepo.UpdateSessionState(roomID, domain.SessionStateClosed)
}
