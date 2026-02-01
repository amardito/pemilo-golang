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

// CreateCandidateRequest represents candidate creation payload
type CreateCandidateRequest struct {
	Name              string `json:"name" binding:"required,min=1,max=255"`
	PhotoURL          string `json:"photo_url"`
	Description       string `json:"description"`
	DisplayOrder      int    `json:"display_order"`
	ParentCandidateID *uint  `json:"parent_candidate_id"`
}

// UpdateCandidateRequest represents candidate update payload
type UpdateCandidateRequest struct {
	Name              string `json:"name" binding:"omitempty,min=1,max=255"`
	PhotoURL          string `json:"photo_url"`
	Description       string `json:"description"`
	DisplayOrder      *int   `json:"display_order"`
	ParentCandidateID *uint  `json:"parent_candidate_id"`
}

// verifyRoomOwnership checks if the user owns the room
func verifyRoomOwnership(c *gin.Context, roomID uint) (*models.VotingRoom, bool) {
	userID := middleware.GetUserID(c)

	var room models.VotingRoom
	if err := config.DB.Where("id = ? AND created_by = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return nil, false
	}

	return &room, true
}

// ListCandidates returns all candidates for a room
func ListCandidates(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	if _, ok := verifyRoomOwnership(c, uint(roomID)); !ok {
		return
	}

	var candidates []models.Candidate
	if err := config.DB.Where("room_id = ?", roomID).
		Order("display_order ASC").
		Find(&candidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch candidates",
		})
		return
	}

	// Add vote counts
	for i := range candidates {
		var count int64
		config.DB.Model(&models.Vote{}).Where("candidate_id = ?", candidates[i].ID).Count(&count)
		candidates[i].VoteCount = int(count)
	}

	c.JSON(http.StatusOK, gin.H{
		"candidates": candidates,
	})
}

// GetCandidate returns a specific candidate
func GetCandidate(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	candidateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid candidate ID",
		})
		return
	}

	if _, ok := verifyRoomOwnership(c, uint(roomID)); !ok {
		return
	}

	var candidate models.Candidate
	if err := config.DB.Where("id = ? AND room_id = ?", candidateID, roomID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		First(&candidate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Candidate not found",
		})
		return
	}

	// Add vote count
	var count int64
	config.DB.Model(&models.Vote{}).Where("candidate_id = ?", candidate.ID).Count(&count)
	candidate.VoteCount = int(count)

	c.JSON(http.StatusOK, gin.H{
		"candidate": candidate,
	})
}

// CreateCandidate creates a new candidate in a room
func CreateCandidate(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	room, ok := verifyRoomOwnership(c, uint(roomID))
	if !ok {
		return
	}

	// Don't allow adding candidates to finished rooms
	if room.Status == models.RoomStatusFinished {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot add candidates to a finished room",
		})
		return
	}

	var req CreateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	candidate := models.Candidate{
		RoomID:            uint(roomID),
		Name:              req.Name,
		PhotoURL:          req.PhotoURL,
		Description:       req.Description,
		DisplayOrder:      req.DisplayOrder,
		ParentCandidateID: req.ParentCandidateID,
	}

	if err := config.DB.Create(&candidate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create candidate",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Candidate created successfully",
		"candidate": candidate,
	})
}

// UpdateCandidate updates an existing candidate
func UpdateCandidate(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	candidateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid candidate ID",
		})
		return
	}

	if _, ok := verifyRoomOwnership(c, uint(roomID)); !ok {
		return
	}

	var candidate models.Candidate
	if err := config.DB.Where("id = ? AND room_id = ?", candidateID, roomID).First(&candidate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Candidate not found",
		})
		return
	}

	var req UpdateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update fields
	if req.Name != "" {
		candidate.Name = req.Name
	}
	if req.PhotoURL != "" {
		candidate.PhotoURL = req.PhotoURL
	}
	if req.Description != "" {
		candidate.Description = req.Description
	}
	if req.DisplayOrder != nil {
		candidate.DisplayOrder = *req.DisplayOrder
	}
	if req.ParentCandidateID != nil {
		candidate.ParentCandidateID = req.ParentCandidateID
	}

	if err := config.DB.Save(&candidate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update candidate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Candidate updated successfully",
		"candidate": candidate,
	})
}

// DeleteCandidate deletes a candidate
func DeleteCandidate(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	candidateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid candidate ID",
		})
		return
	}

	room, ok := verifyRoomOwnership(c, uint(roomID))
	if !ok {
		return
	}

	// Don't allow deleting candidates from active or finished rooms
	if room.Status != models.RoomStatusInactive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete candidates from an active or finished room",
		})
		return
	}

	var candidate models.Candidate
	if err := config.DB.Where("id = ? AND room_id = ?", candidateID, roomID).First(&candidate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Candidate not found",
		})
		return
	}

	if err := config.DB.Delete(&candidate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete candidate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Candidate deleted successfully",
	})
}
