package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketUsecase *usecase.TicketUsecase
}

func NewTicketHandler(ticketUsecase *usecase.TicketUsecase) *TicketHandler {
	return &TicketHandler{
		ticketUsecase: ticketUsecase,
	}
}

// CreateTicket creates a single ticket
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req dto.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ticket, err := h.ticketUsecase.CreateTicket(req.RoomID, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toTicketResponse(ticket))
}

// CreateTicketsBulk creates multiple tickets from a list
func (h *TicketHandler) CreateTicketsBulk(c *gin.Context) {
	var req dto.CreateTicketsBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	tickets, err := h.ticketUsecase.CreateTicketsBulk(req.RoomID, req.Codes)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.ListTicketsResponse{
		Tickets: make([]*dto.TicketResponse, 0, len(tickets)),
	}
	for _, ticket := range tickets {
		response.Tickets = append(response.Tickets, toTicketResponse(ticket))
	}

	c.JSON(http.StatusCreated, response)
}

// ListTicketsByRoom lists all tickets for a room
func (h *TicketHandler) ListTicketsByRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	tickets, err := h.ticketUsecase.GetTicketsByRoom(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.ListTicketsResponse{
		Tickets: make([]*dto.TicketResponse, 0, len(tickets)),
	}
	for _, ticket := range tickets {
		response.Tickets = append(response.Tickets, toTicketResponse(ticket))
	}

	c.JSON(http.StatusOK, response)
}

// DeleteTicket deletes a ticket
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	ticketID := c.Param("id")

	if err := h.ticketUsecase.DeleteTicket(ticketID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Success: true, Message: "Ticket deleted successfully"})
}

// Helper function
func toTicketResponse(ticket *domain.Ticket) *dto.TicketResponse {
	return &dto.TicketResponse{
		ID:        ticket.ID,
		RoomID:    ticket.RoomID,
		Code:      ticket.Code,
		IsUsed:    ticket.IsUsed,
		UsedAt:    ticket.UsedAt,
		CreatedAt: ticket.CreatedAt,
		UpdatedAt: ticket.UpdatedAt,
	}
}
