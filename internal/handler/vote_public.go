package handler

import (
	"net/http"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type VotePublicHandler struct {
	voteService *service.VoteService
}

func NewVotePublicHandler(voteService *service.VoteService) *VotePublicHandler {
	return &VotePublicHandler{voteService: voteService}
}

// POST /api/public/events/:eventId/vote/prepare
func (h *VotePublicHandler) Prepare(c *gin.Context) {
	var req dto.VotePrepareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: "token and nim are required"})
		return
	}

	eventID := c.Param("eventId")

	resp, err := h.voteService.Prepare(c.Request.Context(), eventID, req)
	if err != nil {
		_ = c.Error(err)
		status := mapVoteError(err)
		// Use generic error message — don't leak info
		msg := "invalid token or NIM"
		if err == service.ErrEventNotOpen {
			msg = "voting is not open"
		} else if err == service.ErrAlreadyVoted {
			msg = "you have already voted"
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// POST /api/public/events/:eventId/vote/submit
func (h *VotePublicHandler) Submit(c *gin.Context) {
	var req dto.VoteSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: "token, nim, and slate_id are required"})
		return
	}

	eventID := c.Param("eventId")

	if err := h.voteService.Submit(c.Request.Context(), eventID, req); err != nil {
		_ = c.Error(err)
		status := mapVoteError(err)
		msg := "invalid token or NIM"
		if err == service.ErrEventNotOpen {
			msg = "voting is not open"
		} else if err == service.ErrAlreadyVoted {
			msg = "you have already voted"
		} else if err == service.ErrInvalidSlate {
			msg = "invalid slate selection"
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: msg})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "vote submitted successfully"})
}

func mapVoteError(err error) int {
	switch err {
	case service.ErrEventNotFound:
		return http.StatusNotFound
	case service.ErrEventNotOpen:
		return http.StatusForbidden
	case service.ErrInvalidToken, service.ErrVoterNotEligible:
		return http.StatusUnauthorized
	case service.ErrAlreadyVoted:
		return http.StatusConflict
	case service.ErrInvalidSlate:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
