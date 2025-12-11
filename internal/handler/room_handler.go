package handler

import (
	"net/http"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/internal/dto"
	"github.com/amardito/pemilo-golang/internal/usecase"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomUsecase *usecase.RoomUsecase
}

func NewRoomHandler(roomUsecase *usecase.RoomUsecase) *RoomHandler {
	return &RoomHandler{
		roomUsecase: roomUsecase,
	}
}

// CreateRoom creates a new election room
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	// Get admin ID from JWT context
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Admin ID not found in context"})
		return
	}

	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	room, err := h.roomUsecase.CreateRoom(
		adminID.(string),
		req.Name,
		domain.VotersType(req.VotersType),
		req.VotersLimit,
		req.SessionStartTime,
		req.SessionEndTime,
		domain.RoomStatus(req.Status),
		domain.PublishState(req.PublishState),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toRoomResponse(room))
}

// GetRoom retrieves a room by ID
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")

	room, err := h.roomUsecase.GetRoom(roomID)
	if err != nil {
		if err == domain.ErrRoomNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toRoomResponse(room))
}

// UpdateRoom updates an existing room
func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	roomID := c.Param("id")

	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	var votersType *domain.VotersType
	if req.VotersType != nil {
		vt := domain.VotersType(*req.VotersType)
		votersType = &vt
	}

	var status *domain.RoomStatus
	if req.Status != nil {
		s := domain.RoomStatus(*req.Status)
		status = &s
	}

	var publishState *domain.PublishState
	if req.PublishState != nil {
		ps := domain.PublishState(*req.PublishState)
		publishState = &ps
	}

	room, err := h.roomUsecase.UpdateRoom(
		roomID,
		req.Name,
		votersType,
		req.VotersLimit,
		req.SessionStartTime,
		req.SessionEndTime,
		status,
		publishState,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toRoomResponse(room))
}

// DeleteRoom deletes a room
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")

	if err := h.roomUsecase.DeleteRoom(roomID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Success: true, Message: "Room deleted successfully"})
}

// BulkDeleteRooms deletes multiple rooms
func (h *RoomHandler) BulkDeleteRooms(c *gin.Context) {
	var req dto.BulkDeleteRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Room IDs cannot be empty"})
		return
	}

	if err := h.roomUsecase.BulkDeleteRooms(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Success: true, Message: "Rooms deleted successfully"})
}

// ListRooms lists all rooms with optional filters
func (h *RoomHandler) ListRooms(c *gin.Context) {
	var req dto.ListRoomsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	filters := domain.RoomFilters{}
	if req.Status != nil {
		s := domain.RoomStatus(*req.Status)
		filters.Status = &s
	}
	if req.PublishState != nil {
		ps := domain.PublishState(*req.PublishState)
		filters.PublishState = &ps
	}
	if req.SessionState != nil {
		ss := domain.SessionState(*req.SessionState)
		filters.SessionState = &ss
	}

	rooms, err := h.roomUsecase.ListRooms(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := &dto.ListRoomsResponse{
		Rooms: make([]*dto.RoomResponse, 0, len(rooms)),
	}
	for _, room := range rooms {
		response.Rooms = append(response.Rooms, toRoomResponse(room))
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to convert domain.Room to dto.RoomResponse
func toRoomResponse(room *domain.Room) *dto.RoomResponse {
	return &dto.RoomResponse{
		ID:               room.ID,
		AdminID:          room.AdminID,
		Name:             room.Name,
		VotersType:       string(room.VotersType),
		VotersLimit:      room.VotersLimit,
		SessionStartTime: room.SessionStartTime,
		SessionEndTime:   room.SessionEndTime,
		Status:           string(room.Status),
		PublishState:     string(room.PublishState),
		SessionState:     string(room.SessionState),
		CreatedAt:        room.CreatedAt,
		UpdatedAt:        room.UpdatedAt,
	}
}
