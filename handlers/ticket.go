package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"pemilo/config"
	"pemilo/models"

	"github.com/gin-gonic/gin"
)

// GenerateTicketsRequest represents ticket generation payload
type GenerateTicketsRequest struct {
	Count      int      `json:"count" binding:"required,min=1,max=1000"`
	VoterNames []string `json:"voter_names"` // Optional: pre-assign names
}

// ListTickets returns all tickets for a room
func ListTickets(c *gin.Context) {
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

	var tickets []models.Ticket
	if err := config.DB.Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tickets",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
	})
}

// GenerateTickets creates new tickets for a room
func GenerateTickets(c *gin.Context) {
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

	// Don't allow generating tickets for finished rooms
	if room.Status == models.RoomStatusFinished {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot generate tickets for a finished room",
		})
		return
	}

	var req GenerateTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	tickets := make([]models.Ticket, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		code, err := models.GenerateTicketCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate ticket code",
			})
			return
		}

		ticket := models.Ticket{
			Code:   code,
			RoomID: uint(roomID),
			Status: models.TicketStatusUnused,
		}

		// Assign voter name if provided
		if len(req.VoterNames) > i {
			name := req.VoterNames[i]
			ticket.VoterName = &name
		}

		tickets = append(tickets, ticket)
	}

	if err := config.DB.Create(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create tickets",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("%d tickets created successfully", len(tickets)),
		"tickets": tickets,
	})
}

// ExportTicketsCSV exports tickets as CSV
func ExportTicketsCSV(c *gin.Context) {
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

	var tickets []models.Ticket
	if err := config.DB.Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tickets",
		})
		return
	}

	// Set CSV headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=tickets_%s_%d.csv", sanitizeFilename(room.Name), roomID))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Code", "Voter Name", "Status", "Used At"})

	// Write data
	for _, ticket := range tickets {
		voterName := ""
		if ticket.VoterName != nil {
			voterName = sanitizeCSVField(*ticket.VoterName)
		}
		usedAt := ""
		if ticket.UsedAt != nil {
			usedAt = ticket.UsedAt.Format("2006-01-02 15:04:05")
		}
		writer.Write([]string{
			ticket.Code,
			voterName,
			string(ticket.Status),
			usedAt,
		})
	}
}

// DeleteTicket deletes an unused ticket
func DeleteTicket(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
		})
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ticket ID",
		})
		return
	}

	if _, ok := verifyRoomOwnership(c, uint(roomID)); !ok {
		return
	}

	var ticket models.Ticket
	if err := config.DB.Where("id = ? AND room_id = ?", ticketID, roomID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ticket not found",
		})
		return
	}

	// Don't allow deleting used tickets
	if ticket.IsUsed() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete a used ticket",
		})
		return
	}

	if err := config.DB.Delete(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete ticket",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket deleted successfully",
	})
}

// sanitizeFilename removes unsafe characters from filename
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	return string(result)
}

// sanitizeCSVField sanitizes field for CSV output (prevent injection)
func sanitizeCSVField(field string) string {
	if len(field) > 0 {
		first := field[0]
		// Prefix with single quote if starts with potentially dangerous characters
		if first == '=' || first == '+' || first == '-' || first == '@' || first == '\t' || first == '\r' {
			return "'" + field
		}
	}
	return field
}
