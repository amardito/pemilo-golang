package handler

import (
	"net/http"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == service.ErrEmailTaken {
			c.JSON(http.StatusConflict, dto.ErrorResponse{OK: false, Error: err.Error()})
			return
		}
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{OK: true, Data: resp})
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{OK: false, Error: "invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: resp})
}

// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authService.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{OK: false, Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: user})
}

// POST /api/auth/logout — stateless, just acknowledge
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "logged out"})
}
