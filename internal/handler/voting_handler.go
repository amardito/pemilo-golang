package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type VotingHandler struct {
	votingUsecase *usecase.VotingUsecase
}

func NewVotingHandler(votingUsecase *usecase.VotingUsecase) *VotingHandler {
	return &VotingHandler{
		votingUsecase: votingUsecase,
	}
}

// GetVoterRoomInfo returns room information for voters
func (h *VotingHandler) GetVoterRoomInfo(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "room_id is required"})
		return
	}

	room, candidates, requiresTicket, message, err := h.votingUsecase.GetVoterRoomInfo(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Message: message})
		return
	}

	// Convert candidates
	candidatesResponse := make([]*dto.CandidateResponse, 0, len(candidates))
	for _, candidate := range candidates {
		candidatesResponse = append(candidatesResponse, toCandidateResponseFromDomain(candidate))
	}

	response := &dto.GetVoterRoomInfoResponse{
		Room:           *toRoomResponse(room),
		Candidates:     candidatesResponse,
		RequiresTicket: requiresTicket,
		IsActive:       true,
		Message:        message,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyTicket verifies a ticket code
func (h *VotingHandler) VerifyTicket(c *gin.Context) {
	var req dto.VerifyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Simple verification - actual validation happens in CastVote
	c.JSON(http.StatusOK, dto.VerifyTicketResponse{
		Valid:   true,
		Message: "Proceed to vote",
	})
}

// CastVote handles vote submission
func (h *VotingHandler) CastVote(c *gin.Context) {
	var req dto.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	vote, err := h.votingUsecase.CastVote(req.RoomID, req.CandidateID, req.SubCandidateID, req.TicketCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.VoteResponse{
		ID:              vote.ID,
		RoomID:          vote.RoomID,
		CandidateID:     vote.CandidateID,
		SubCandidateID:  vote.SubCandidateID,
		VoterIdentifier: vote.VoterIdentifier,
		CreatedAt:       vote.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetRealtimeVoteData returns real-time voting statistics
func (h *VotingHandler) GetRealtimeVoteData(c *gin.Context) {
	roomID := c.Param("roomId")

	room, voteCounts, totalVotes, err := h.votingUsecase.GetRealtimeVoteData(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	voteData := make([]*dto.RealtimeVoteData, 0, len(voteCounts))
	for _, vc := range voteCounts {
		voteData = append(voteData, &dto.RealtimeVoteData{
			CandidateID:   vc.CandidateID,
			CandidateName: "", // TODO: fetch candidate name
			VoteCount:     vc.Count,
			Timestamp:     vc.Timestamp,
		})
	}

	response := &dto.RealtimeVoteResponse{
		RoomID:     room.ID,
		RoomName:   room.Name,
		VoteData:   voteData,
		TotalVotes: totalVotes,
		UpdatedAt:  voteCounts[0].Timestamp,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions
func toCandidateResponseFromDomain(candidate *domain.CandidateWithSubs) *dto.CandidateResponse {
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
