package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/middleware"
	"github.com/amard/pemilo-golang/internal/service"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// POST /api/events/:eventId/upgrade
func (h *PaymentHandler) Upgrade(c *gin.Context) {
	var req dto.UpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	resp, err := h.paymentService.Upgrade(c.Request.Context(), eventID, userID, req)
	if err != nil {
		_ = c.Error(err)
		status := http.StatusInternalServerError
		switch err {
		case service.ErrEventNotFound:
			status = http.StatusNotFound
		case service.ErrEventForbidden:
			status = http.StatusForbidden
		case service.ErrAlreadyUpgraded:
			status = http.StatusConflict
		case service.ErrPaymentFailed:
			status = http.StatusBadGateway
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: resp})
}

// POST /api/payments/ipaymu/webhook
func (h *PaymentHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{OK: false, Error: "invalid body"})
		return
	}

	signature := c.GetHeader("X-Signature")

	if err := h.paymentService.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		if err == service.ErrInvalidSignature {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{OK: false, Error: "invalid signature"})
			return
		}
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{OK: false, Error: "webhook processing failed"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Message: "webhook processed"})
}

// GET /api/orders/:orderId
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	orderID := c.Param("orderId")

	order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		_ = c.Error(err)
		status := http.StatusInternalServerError
		if err == service.ErrOrderNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrEventForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, dto.ErrorResponse{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: order})
}

// GET /api/events/:eventId/audit-logs
type AuditLogHandler struct {
	auditService *service.AuditService
}

func NewAuditLogHandler(auditService *service.AuditService) *AuditLogHandler {
	return &AuditLogHandler{auditService: auditService}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	eventID := c.Param("eventId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	resp, err := h.auditService.List(c.Request.Context(), eventID, userID, page, perPage)
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

	c.JSON(http.StatusOK, dto.SuccessResponse{OK: true, Data: resp})
}
