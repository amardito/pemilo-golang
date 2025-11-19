package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminUsecase *usecase.AdminUsecase
}

func NewAdminHandler(adminUsecase *usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{
		adminUsecase: adminUsecase,
	}
}

// CreateAdmin creates new admin account (Basic Auth protected, owner only)
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	admin, err := h.adminUsecase.CreateAdmin(
		req.Username,
		req.EncryptedPassword,
		req.MaxRoom,
		req.MaxVoters,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.CreateAdminResponse{
		ID:        admin.ID,
		Username:  admin.Username,
		MaxRoom:   admin.MaxRoom,
		MaxVoters: admin.MaxVoters,
		IsActive:  admin.IsActive,
		CreatedAt: admin.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetAdminQuota retrieves admin's quota usage
func (h *AdminHandler) GetAdminQuota(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Admin ID not found in context"})
		return
	}

	admin, currentRooms, currentVoters, err := h.adminUsecase.GetAdminQuota(adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.GetAdminQuotaResponse{
		Admin: dto.AdminInfo{
			ID:        admin.ID,
			Username:  admin.Username,
			MaxRoom:   admin.MaxRoom,
			MaxVoters: admin.MaxVoters,
			IsActive:  admin.IsActive,
		},
		CurrentRooms:  currentRooms,
		CurrentVoters: currentVoters,
		RoomLimit:     admin.MaxRoom,
		VotersLimit:   admin.MaxVoters,
	}

	c.JSON(http.StatusOK, response)
}
