package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type CandidateHandler struct {
	candidateUsecase *usecase.CandidateUsecase
}

func NewCandidateHandler(candidateUsecase *usecase.CandidateUsecase) *CandidateHandler {
	return &CandidateHandler{
		candidateUsecase: candidateUsecase,
	}
}

// CreateCandidate creates a new candidate with optional sub-candidates
func (h *CandidateHandler) CreateCandidate(c *gin.Context) {
	var req dto.CreateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert sub-candidates
	subCandidates := make([]struct {
		Name        string
		PhotoURL    string
		Description string
	}, 0, len(req.SubCandidates))

	for _, sub := range req.SubCandidates {
		subCandidates = append(subCandidates, struct {
			Name        string
			PhotoURL    string
			Description string
		}{
			Name:        sub.Name,
			PhotoURL:    sub.PhotoURL,
			Description: sub.Description,
		})
	}

	candidate, err := h.candidateUsecase.CreateCandidate(
		req.RoomID,
		req.Name,
		req.PhotoURL,
		req.Description,
		subCandidates,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toCandidateResponse(candidate))
}

// GetCandidate retrieves a candidate by ID
func (h *CandidateHandler) GetCandidate(c *gin.Context) {
	candidateID := c.Param("id")

	candidate, err := h.candidateUsecase.GetCandidate(candidateID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Candidate not found"})
		return
	}

	c.JSON(http.StatusOK, toCandidateResponse(candidate))
}

// ListCandidatesByRoom lists all candidates for a specific room
func (h *CandidateHandler) ListCandidatesByRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	candidates, err := h.candidateUsecase.GetCandidatesByRoom(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.ListCandidatesResponse{
		Candidates: make([]*dto.CandidateResponse, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		response.Candidates = append(response.Candidates, toCandidateResponse(candidate))
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCandidate updates a candidate
func (h *CandidateHandler) UpdateCandidate(c *gin.Context) {
	candidateID := c.Param("id")

	var req dto.UpdateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	candidate, err := h.candidateUsecase.UpdateCandidate(candidateID, req.Name, req.PhotoURL, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}

// DeleteCandidate deletes a candidate
func (h *CandidateHandler) DeleteCandidate(c *gin.Context) {
	candidateID := c.Param("id")

	if err := h.candidateUsecase.DeleteCandidate(candidateID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Success: true, Message: "Candidate deleted successfully"})
}

// Helper function
func toCandidateResponse(candidate *domain.CandidateWithSubs) *dto.CandidateResponse {
	response := &dto.CandidateResponse{
		ID:          candidate.Candidate.ID,
		RoomID:      candidate.Candidate.RoomID,
		Name:        candidate.Candidate.Name,
		PhotoURL:    candidate.Candidate.PhotoURL,
		Description: candidate.Candidate.Description,
		CreatedAt:   candidate.Candidate.CreatedAt,
		UpdatedAt:   candidate.Candidate.UpdatedAt,
	}

	response.SubCandidates = make([]dto.SubCandidateResponse, 0, len(candidate.SubCandidates))
	for _, sub := range candidate.SubCandidates {
		response.SubCandidates = append(response.SubCandidates, dto.SubCandidateResponse{
			ID:          sub.ID,
			CandidateID: sub.CandidateID,
			Name:        sub.Name,
			PhotoURL:    sub.PhotoURL,
			Description: sub.Description,
			CreatedAt:   sub.CreatedAt,
			UpdatedAt:   sub.UpdatedAt,
		})
	}

	return response
}
