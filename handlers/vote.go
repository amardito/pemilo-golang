package handlers

import (
	"net/http"
	"strconv"

	"pemilo/config"
	"pemilo/models"

	"github.com/gin-gonic/gin"
)

// CastVoteRequest represents vote casting payload
type CastVoteRequest struct {
	TicketCode  string `json:"ticket_code" binding:"required"`
	CandidateID uint   `json:"candidate_id" binding:"required"`
}

// ValidateTicket checks if a ticket is valid for voting
func ValidateTicket(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ticket code is required",
		})
		return
	}

	var ticket models.Ticket
	if err := config.DB.Where("code = ?", code).
		Preload("Room").
		First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid": false,
			"error": "Ticket not found",
		})
		return
	}

	// Check if ticket is already used
	if ticket.IsUsed() {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Ticket has already been used",
		})
		return
	}

	// Check if room is active
	if !ticket.Room.CanVote() {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Voting is not currently active for this room",
		})
		return
	}

	// Get candidates for the room
	var candidates []models.Candidate
	config.DB.Where("room_id = ?", ticket.RoomID).
		Order("display_order ASC").
		Find(&candidates)

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"room":       ticket.Room,
		"candidates": candidates,
	})
}

// CastVote handles vote submission
func CastVote(c *gin.Context) {
	var req CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Start transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock ticket row for update (prevent race conditions)
	var ticket models.Ticket
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("code = ?", req.TicketCode).
		Preload("Room").
		First(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ticket not found",
		})
		return
	}

	// Check if ticket is already used
	if ticket.IsUsed() {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ticket has already been used",
		})
		return
	}

	// Check if room is active
	if !ticket.Room.CanVote() {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Voting is not currently active for this room",
		})
		return
	}

	// Verify candidate belongs to the room
	var candidate models.Candidate
	if err := tx.Where("id = ? AND room_id = ?", req.CandidateID, ticket.RoomID).First(&candidate).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid candidate for this room",
		})
		return
	}

	// Create vote
	vote := models.Vote{
		TicketID:    ticket.ID,
		CandidateID: req.CandidateID,
	}

	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record vote",
		})
		return
	}

	// Mark ticket as used
	ticket.MarkUsed()
	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update ticket",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to complete vote",
		})
		return
	}

	// Broadcast vote update via WebSocket (if hub is available)
	BroadcastVoteUpdate(ticket.RoomID, req.CandidateID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vote recorded successfully",
	})
}

// GetRoomResults returns voting results for a room (admin only)
func GetRoomResults(c *gin.Context) {
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

	// Get candidates with vote counts
	var candidates []models.Candidate
	if err := config.DB.Where("room_id = ?", roomID).
		Order("display_order ASC").
		Find(&candidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch candidates",
		})
		return
	}

	type CandidateResult struct {
		models.Candidate
		VoteCount  int     `json:"vote_count"`
		Percentage float64 `json:"percentage"`
	}

	// Calculate total votes
	var totalVotes int64
	config.DB.Model(&models.Vote{}).
		Joins("JOIN tickets ON votes.ticket_id = tickets.id").
		Where("tickets.room_id = ?", roomID).
		Count(&totalVotes)

	results := make([]CandidateResult, len(candidates))
	for i, candidate := range candidates {
		var count int64
		config.DB.Model(&models.Vote{}).Where("candidate_id = ?", candidate.ID).Count(&count)

		percentage := 0.0
		if totalVotes > 0 {
			percentage = float64(count) / float64(totalVotes) * 100
		}

		results[i] = CandidateResult{
			Candidate:  candidate,
			VoteCount:  int(count),
			Percentage: percentage,
		}
		results[i].Candidate.VoteCount = int(count)
	}

	// Get ticket stats
	var totalTickets, usedTickets int64
	config.DB.Model(&models.Ticket{}).Where("room_id = ?", roomID).Count(&totalTickets)
	config.DB.Model(&models.Ticket{}).Where("room_id = ? AND status = ?", roomID, models.TicketStatusUsed).Count(&usedTickets)

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"total_votes":   totalVotes,
		"total_tickets": totalTickets,
		"used_tickets":  usedTickets,
		"participation": func() float64 {
			if totalTickets > 0 {
				return float64(usedTickets) / float64(totalTickets) * 100
			}
			return 0
		}(),
	})
}
