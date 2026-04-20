package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type VoterHandler struct {
	voterService *service.VoterService
}

func NewVoterHandler(voterService *service.VoterService) *VoterHandler {
	return &VoterHandler{voterService: voterService}
}

// POST /api/events/:eventId/voters/import
func (h *VoterHandler) Import(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: "file is required"})
		return
	}
	defer file.Close()

	result, err := h.voterService.ImportCSV(c.Request.Context(), eventID, userID, file)
	if err != nil {
		_ = c.Error(err)
		status := http.StatusInternalServerError
		if err == service.ErrEventNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrEventForbidden {
			status = http.StatusForbidden
		} else if err == service.ErrEventLocked {
			status = http.StatusConflict
		} else if err == service.ErrMaxVotersReached {
			status = http.StatusBadRequest
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: result})
}

// GET /api/events/:eventId/voters
func (h *VoterHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	var hasVoted *bool
	if hv := c.Query("has_voted"); hv != "" {
		v := hv == "true"
		hasVoted = &v
	}

	params := dto.VoterListParams{
		Query:    c.Query("q"),
		Status:   c.Query("status"),
		HasVoted: hasVoted,
		Page:     page,
		PerPage:  perPage,
	}

	result, err := h.voterService.List(c.Request.Context(), eventID, userID, params)
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

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: result})
}

// POST /api/events/:eventId/voters/tokens/generate
func (h *VoterHandler) GenerateTokens(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	count, err := h.voterService.GenerateTokens(c.Request.Context(), eventID, userID)
	if err != nil {
		_ = c.Error(err)
		status := http.StatusInternalServerError
		if err == service.ErrEventNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrEventForbidden {
			status = http.StatusForbidden
		} else if err == service.ErrEventLocked {
			status = http.StatusConflict
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: gin.H{"generated_count": count}})
}

// GET /api/events/:eventId/voters/tokens/export
func (h *VoterHandler) ExportTokens(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	rows, err := h.voterService.ExportTokens(c.Request.Context(), eventID, userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=tokens_%s.csv", eventID))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"full_name", "nim", "class_name", "token"})
	for _, row := range rows {
		className := ""
		if row.ClassName != nil {
			className = *row.ClassName
		}
		w.Write([]string{row.FullName, row.NIMRaw, className, row.Token})
	}
	w.Flush()
}

// GET /api/events/:eventId/voters/turnout/export
func (h *VoterHandler) ExportTurnout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	voters, err := h.voterService.ExportTurnout(c.Request.Context(), eventID, userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=turnout_%s.csv", eventID))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"full_name", "nim", "class_name", "has_voted", "voted_at"})
	for _, v := range voters {
		className := ""
		if v.ClassName != nil {
			className = *v.ClassName
		}
		votedAt := ""
		if v.VotedAt != nil {
			votedAt = v.VotedAt.Format("2006-01-02 15:04:05")
		}
		w.Write([]string{v.FullName, v.NIMRaw, className, fmt.Sprintf("%t", v.HasVoted), votedAt})
	}
	w.Flush()
}

// GET /api/voters/template
func (h *VoterHandler) DownloadTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=template_pemilih.csv")

	w := csv.NewWriter(c.Writer)
	// Header row
	w.Write([]string{"full_name", "nim", "class_name"})
	// Example rows so users understand the expected format
	w.Write([]string{"Budi Santoso", "2023010001", "TI-A"})
	w.Write([]string{"Siti Rahayu", "2023010002", "TI-B"})
	w.Write([]string{"Andi Wijaya", "2023010003", ""})
	w.Flush()
}
