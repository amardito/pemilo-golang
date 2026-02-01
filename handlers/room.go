package handlers

import (
	"net/http"
	"strconv"

	"pemilo/config"
	"pemilo/middleware"
	"pemilo/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateRoomRequest represents room creation payload
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
}

// UpdateRoomRequest represents room update payload
type UpdateRoomRequest struct {
	Name        string            `json:"name" binding:"omitempty,min=1,max=255"`
	Description string            `json:"description"`
	Status      models.RoomStatus `json:"status" binding:"omitempty,oneof=inactive active finished"`
}

// ListRooms returns all voting rooms for the authenticated user
func ListRooms(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var rooms []models.VotingRoom
	if err := config.DB.Where("created_by = ?", userID).
		Preload("Candidates").
		Order("created_at DESC").
		Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rooms",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// GetRoom returns a specific voting room
func GetRoom(c *gin.Context) {
	userID := middleware.GetUserID(c)
	roomID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	var room models.VotingRoom
	if err := config.DB.Where("id = ? AND created_by = ?", roomID, userID).
		Preload("Candidates", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return
	}

	// Count votes for each candidate
	for i := range room.Candidates {
		var count int64
		config.DB.Model(&models.Vote{}).Where("candidate_id = ?", room.Candidates[i].ID).Count(&count)
		room.Candidates[i].VoteCount = int(count)
	}

	// Count tickets
	var totalTickets, usedTickets int64
	config.DB.Model(&models.Ticket{}).Where("room_id = ?", roomID).Count(&totalTickets)
	config.DB.Model(&models.Ticket{}).Where("room_id = ? AND status = ?", roomID, models.TicketStatusUsed).Count(&usedTickets)

	c.JSON(http.StatusOK, gin.H{
		"room":          room,
		"total_tickets": totalTickets,
		"used_tickets":  usedTickets,
	})
}

// CreateRoom creates a new voting room
func CreateRoom(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	room := models.VotingRoom{
		Name:        req.Name,
		Description: req.Description,
		Status:      models.RoomStatusInactive,
		CreatedBy:   userID,
	}

	if err := config.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create room",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"room":    room,
	})
}

// UpdateRoom updates an existing voting room
func UpdateRoom(c *gin.Context) {
	userID := middleware.GetUserID(c)
	roomID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	var room models.VotingRoom
	if err := config.DB.Where("id = ? AND created_by = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return
	}

	var req UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update fields
	if req.Name != "" {
		room.Name = req.Name
	}
	if req.Description != "" {
		room.Description = req.Description
	}
	if req.Status != "" {
		room.Status = req.Status
	}

	if err := config.DB.Save(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Room updated successfully",
		"room":    room,
	})
}

// DeleteRoom deletes a voting room
func DeleteRoom(c *gin.Context) {
	userID := middleware.GetUserID(c)
	roomID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	var room models.VotingRoom
	if err := config.DB.Where("id = ? AND created_by = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return
	}

	// Soft delete room and related data
	if err := config.DB.Delete(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Room deleted successfully",
	})
}
