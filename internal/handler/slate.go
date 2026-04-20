package handler

import (
	"net/http"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type SlateHandler struct {
	slateService *service.SlateService
}

func NewSlateHandler(slateService *service.SlateService) *SlateHandler {
	return &SlateHandler{slateService: slateService}
}

// POST /api/events/:eventId/slates
func (h *SlateHandler) Create(c *gin.Context) {
	var req dto.CreateSlateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	slate, err := h.slateService.Create(c.Request.Context(), eventID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{OK: true, Data: slate})
}

// GET /api/events/:eventId/slates
func (h *SlateHandler) List(c *gin.Context) {
	eventID := c.Param("eventId")

	slates, err := h.slateService.List(c.Request.Context(), eventID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: "failed to list slates"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: slates})
}

// PATCH /api/slates/:slateId
func (h *SlateHandler) Update(c *gin.Context) {
	var req dto.UpdateSlateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	slateID := c.Param("slateId")

	slate, err := h.slateService.Update(c.Request.Context(), slateID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: slate})
}

// DELETE /api/slates/:slateId
func (h *SlateHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	slateID := c.Param("slateId")

	if err := h.slateService.Delete(c.Request.Context(), slateID, userID); err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "slate deleted"})
}

// POST /api/slates/:slateId/members
func (h *SlateHandler) CreateMember(c *gin.Context) {
	var req dto.CreateSlateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	slateID := c.Param("slateId")

	member, err := h.slateService.CreateMember(c.Request.Context(), slateID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{OK: true, Data: member})
}

// PATCH /api/slate-members/:memberId
func (h *SlateHandler) UpdateMember(c *gin.Context) {
	var req dto.UpdateSlateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	memberID := c.Param("memberId")

	member, err := h.slateService.UpdateMember(c.Request.Context(), memberID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: member})
}

// DELETE /api/slate-members/:memberId
func (h *SlateHandler) DeleteMember(c *gin.Context) {
	userID := middleware.GetUserID(c)
	memberID := c.Param("memberId")

	if err := h.slateService.DeleteMember(c.Request.Context(), memberID, userID); err != nil {
		_ = c.Error(err)
		status := mapSlateError(err)
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "member deleted"})
}

func mapSlateError(err error) int {
	switch err {
	case service.ErrSlateNotFound, service.ErrMemberNotFound:
		return http.StatusNotFound
	case service.ErrEventForbidden:
		return http.StatusForbidden
	case service.ErrEventLocked, service.ErrSlateHasBallots, service.ErrSlateNotEditable:
		return http.StatusConflict
	case service.ErrMaxSlatesReached:
		return http.StatusBadRequest
	case service.ErrEventNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
