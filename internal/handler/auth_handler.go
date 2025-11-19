package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// Login handles admin authentication (NO Authorization header required)
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	admin, token, expiresAt, err := h.authUsecase.Login(req.Username, req.EncryptedPassword)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err.Error() == "too many failed login attempts" || err == domain.ErrRateLimitExceeded {
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin: dto.AdminInfo{
			ID:        admin.ID,
			Username:  admin.Username,
			MaxRoom:   admin.MaxRoom,
			MaxVoters: admin.MaxVoters,
			IsActive:  admin.IsActive,
		},
	}

	c.JSON(http.StatusOK, response)
}
