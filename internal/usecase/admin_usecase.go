package usecase

import (
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/pkg/utils"
	"github.com/google/uuid"
)

type AdminUsecase struct {
	adminRepo           domain.AdminRepository
	roomRepo            domain.RoomRepository
	encryptionKey       string
	encryptionSaltFront string
	encryptionSaltBack  string
}

func NewAdminUsecase(
	adminRepo domain.AdminRepository,
	roomRepo domain.RoomRepository,
	encryptionKey string,
	encryptionSaltFront string,
	encryptionSaltBack string,
) *AdminUsecase {
	return &AdminUsecase{
		adminRepo:           adminRepo,
		roomRepo:            roomRepo,
		encryptionKey:       encryptionKey,
		encryptionSaltFront: encryptionSaltFront,
		encryptionSaltBack:  encryptionSaltBack,
	}
}

// CreateAdmin creates new admin with plain password (Basic Auth protected endpoint)
func (u *AdminUsecase) CreateAdmin(username, plainPassword string, maxRoom, maxVoters int) (*domain.Admin, error) {
	// Check if admin already exists
	existing, err := u.adminRepo.GetByUsername(username)
	if err == nil && existing != nil {
		return nil, domain.ErrAdminExists
	}

	// Encrypt password first (matching frontend encryption flow)
	encryptedPassword, err := utils.EncryptPassword(plainPassword, u.encryptionKey, u.encryptionSaltFront, u.encryptionSaltBack)
	if err != nil {
		return nil, err
	}

	// Then hash the encrypted password for storage
	hashedPassword, err := utils.HashPassword(encryptedPassword)
	if err != nil {
		return nil, err
	}

	// Create admin
	admin := &domain.Admin{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  hashedPassword,
		MaxRoom:   maxRoom,
		MaxVoters: maxVoters,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.adminRepo.Create(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdmin retrieves admin by ID
func (u *AdminUsecase) GetAdmin(id string) (*domain.Admin, error) {
	return u.adminRepo.GetByID(id)
}

// GetAdminQuota returns admin's current usage vs limits
func (u *AdminUsecase) GetAdminQuota(adminID string) (*domain.Admin, int, int, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, 0, 0, err
	}

	currentRooms, err := u.adminRepo.GetRoomCount(adminID)
	if err != nil {
		return nil, 0, 0, err
	}

	currentVoters, err := u.adminRepo.GetTotalVotersCount(adminID)
	if err != nil {
		return nil, 0, 0, err
	}

	return admin, currentRooms, currentVoters, nil
}

// CheckRoomQuota validates if admin can create a new room
func (u *AdminUsecase) CheckRoomQuota(adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	currentRooms, err := u.adminRepo.GetRoomCount(adminID)
	if err != nil {
		return err
	}

	if currentRooms >= admin.MaxRoom {
		return domain.ErrMaxRoomExceeded
	}

	return nil
}

// CheckVotersQuota validates if admin can add more voters
func (u *AdminUsecase) CheckVotersQuota(adminID string, additionalVoters int) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	currentVoters, err := u.adminRepo.GetTotalVotersCount(adminID)
	if err != nil {
		return err
	}

	if currentVoters+additionalVoters > admin.MaxVoters {
		return domain.ErrMaxVotersExceeded
	}

	return nil
}

// UpdateAdmin updates admin settings
func (u *AdminUsecase) UpdateAdmin(id string, maxRoom, maxVoters *int, isActive *bool) (*domain.Admin, error) {
	admin, err := u.adminRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if maxRoom != nil {
		admin.MaxRoom = *maxRoom
	}
	if maxVoters != nil {
		admin.MaxVoters = *maxVoters
	}
	if isActive != nil {
		admin.IsActive = *isActive
	}

	if err := u.adminRepo.Update(admin); err != nil {
		return nil, err
	}

	return admin, nil
}
