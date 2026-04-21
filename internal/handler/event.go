package handler

import (
	"net/http"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// POST /api/events
func (h *EventHandler) Create(c *gin.Context) {
	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	event, err := h.eventService.Create(c.Request.Context(), userID, req)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{OK: true, Data: event})
}

// GET /api/events
func (h *EventHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	events, err := h.eventService.List(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: "failed to list events"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: events})
}

// GET /api/events/:eventId
func (h *EventHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	event, err := h.eventService.GetByID(c.Request.Context(), eventID, userID)
	if err != nil {
		status := http.StatusNotFound
		if err == service.ErrEventForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: event})
}

// GET /api/public/events/:eventId — no auth required, returns public fields only
func (h *EventHandler) GetPublic(c *gin.Context) {
	eventID := c.Param("eventId")
	info, err := h.eventService.GetPublicInfo(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{OK: false, Error: "event tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: info})
}

// PATCH /api/events/:eventId
func (h *EventHandler) Update(c *gin.Context) {
	var req dto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	event, err := h.eventService.Update(c.Request.Context(), eventID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapEventError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: event})
}

// POST /api/events/:eventId/open
func (h *EventHandler) Open(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	if err := h.eventService.Open(c.Request.Context(), eventID, userID); err != nil {
		_ = c.Error(err)
		status := mapEventError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "event opened"})
}

// POST /api/events/:eventId/close
func (h *EventHandler) Close(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	if err := h.eventService.Close(c.Request.Context(), eventID, userID); err != nil {
		_ = c.Error(err)
		status := mapEventError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "event closed"})
}

// POST /api/events/:eventId/lock
func (h *EventHandler) Lock(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	if err := h.eventService.Lock(c.Request.Context(), eventID, userID); err != nil {
		_ = c.Error(err)
		status := mapEventError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "event locked"})
}

func mapEventError(err error) int {
	switch err {
	case service.ErrEventNotFound:
		return http.StatusNotFound
	case service.ErrEventForbidden:
		return http.StatusForbidden
	case service.ErrEventLocked:
		return http.StatusConflict
	case service.ErrInvalidTransition:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
