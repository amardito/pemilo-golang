package handler

import (
	"net/http"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService *service.StatsService
	jwtSecret    string
}

func NewStatsHandler(statsService *service.StatsService, jwtSecret string) *StatsHandler {
	return &StatsHandler{statsService: statsService, jwtSecret: jwtSecret}
}

// GET /api/events/:eventId/stats
func (h *StatsHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	stats, err := h.statsService.GetStats(c.Request.Context(), eventID, userID)
	if err != nil {
		_ = c.Error(err)
		status := http.StatusInternalServerError
		if err == service.ErrEventNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrEventForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
