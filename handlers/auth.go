package handlers

import (
	"net/http"
	"time"

	"pemilo/config"
	"pemilo/middleware"
	"pemilo/models"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents login payload
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration payload
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Find user by username
	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	config.DB.Save(&user)

	// Create session
	session, err := config.SessionStore.Get(c.Request, middleware.SessionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create session",
		})
		return
	}

	session.Values[middleware.UserIDKey] = user.ID
	session.Values[middleware.UsernameKey] = user.Username

	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save session",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		User:    &user,
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	session, err := config.SessionStore.Get(c.Request, middleware.SessionName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out",
		})
		return
	}

	// Clear session values
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1 // Delete cookie

	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetMe returns current authenticated user
func GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Register handles user registration (for initial setup only)
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Check if username already exists
	var existingUser models.User
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// Create new user
	user := models.User{
		Username: req.Username,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "User created successfully",
		User:    &user,
	})
}
