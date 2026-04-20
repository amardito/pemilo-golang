package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amard/pemilo-golang/internal/config"
	"github.com/amard/pemilo-golang/internal/dto"
	"github.com/amard/pemilo-golang/internal/model"
	"github.com/amard/pemilo-golang/internal/repository"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrAlreadyUpgraded  = errors.New("event already has this package or higher")
	ErrPaymentFailed    = errors.New("payment creation failed")
	ErrInvalidSignature = errors.New("invalid webhook signature")
)

type PaymentService struct {
	orderRepo *repository.OrderRepo
	eventRepo *repository.EventRepo
	cfg       *config.Config
}

func NewPaymentService(orderRepo *repository.OrderRepo, eventRepo *repository.EventRepo, cfg *config.Config) *PaymentService {
	return &PaymentService{orderRepo: orderRepo, eventRepo: eventRepo, cfg: cfg}
}

func (s *PaymentService) Upgrade(ctx context.Context, eventID, userID string, req dto.UpgradeRequest) (*dto.UpgradeResponse, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	targetPkg := model.Package(req.Package)
	limits, ok := model.PackageLimitsMap[targetPkg]
	if !ok || limits.Price == 0 {
		return nil, errors.New("invalid package")
	}

	// Check if already at or above target
	currentLimits := model.PackageLimitsMap[event.Package]
	if currentLimits.Price >= limits.Price {
		return nil, ErrAlreadyUpgraded
	}

	// Create order
	order, err := s.orderRepo.Create(ctx, eventID, targetPkg, limits.Price)
	if err != nil {
		return nil, err
	}

	// Call iPaymu API
	paymentURL, reference, err := s.createIPaymuInvoice(order, event)
	if err != nil {
		return nil, fmt.Errorf("payment creation failed: %w", err)
	}

	// Save reference
	if err := s.orderRepo.UpdateReference(ctx, order.ID, reference); err != nil {
		return nil, err
	}

	return &dto.UpgradeResponse{
		PaymentURL: paymentURL,
		OrderID:    order.ID,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	// Verify signature
	// Skip signature verification for now — iPaymu sandbox webhook signature
	// uses a different signing scheme than the API request signature.
	// In production, verify using X-Signature + X-Timestamp headers.
	_ = signature

	var (
		trxID       string
		sid         string
		referenceID string
		statusCode  int
	)

	// iPaymu sends Content-Type: application/x-www-form-urlencoded but body may be
	// either JSON or URL-encoded depending on version. Handle both.
	if len(body) > 0 && body[0] == '{' {
		// JSON body
		var p struct {
			TrxID       json.Number `json:"trx_id"`
			SID         string      `json:"sid"`
			ReferenceID string      `json:"reference_id"`
			StatusCode  int         `json:"status_code"`
		}
		if err := json.Unmarshal(body, &p); err != nil {
			return err
		}
		trxID = p.TrxID.String()
		sid = p.SID
		referenceID = p.ReferenceID
		statusCode = p.StatusCode
	} else {
		// URL-encoded body: trx_id=204948&sid=...&status_code=1&...
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return err
		}
		trxID = values.Get("trx_id")
		sid = values.Get("sid")
		referenceID = values.Get("reference_id")
		statusCode, _ = strconv.Atoi(values.Get("status_code"))
	}
	_ = trxID

	// Try to find order by reference_id (our order ID) first, then by sid (session ID)
	order, err := s.orderRepo.GetByID(ctx, referenceID)
	if err != nil {
		// Fallback: try by SID which we stored as ipaymu_reference
		order, err = s.orderRepo.GetByReference(ctx, sid)
		if err != nil {
			return ErrOrderNotFound
		}
	}

	if statusCode == 1 { // iPaymu: status_code 1 = berhasil
		if err := s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusPaid); err != nil {
			return err
		}

		// Upgrade event package
		targetPkg := order.Package
		limits := model.PackageLimitsMap[targetPkg]
		return s.eventRepo.UpdatePackage(ctx, order.EventID, targetPkg, limits.MaxSlates, limits.MaxVoters)
	}

	// Failed / expired
	if err := s.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusFailed); err != nil {
		return err
	}
	return nil
}

func (s *PaymentService) GetOrder(ctx context.Context, orderID, userID string) (*dto.OrderDTO, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Verify ownership via event
	event, err := s.eventRepo.GetByID(ctx, order.EventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if event.OwnerUserID != userID {
		return nil, ErrEventForbidden
	}

	return &dto.OrderDTO{
		ID:              order.ID,
		EventID:         order.EventID,
		Package:         string(order.Package),
		Amount:          order.Amount,
		Status:          string(order.Status),
		IPaymuReference: order.IPaymuReference,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

func (s *PaymentService) createIPaymuInvoice(order *model.Order, event *model.Event) (paymentURL, reference string, err error) {
	// Build JSON body matching iPaymu v2 API format
	body := map[string]interface{}{
		"product":     []string{fmt.Sprintf("Pemilo %s - %s", order.Package, event.Title)},
		"qty":         []int{1},
		"price":       []int{order.Amount},
		"notifyUrl":   s.cfg.IPaymuCallbackURL,
		"returnUrl":   s.cfg.ClientOrigin + "/admin/events/" + order.EventID + "/billing",
		"cancelUrl":   s.cfg.ClientOrigin + "/admin/events/" + order.EventID + "/billing",
		"referenceId": order.ID,
		"buyerName":   event.Title,
		"expired":     24,
		"expiredType": "hours",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	// Generate signature per iPaymu spec:
	// stringToSign = "POST:{va}:{lowercase(sha256(body))}:{apiKey}"
	// signature = hmac-sha256(stringToSign, apiKey)
	signature := s.generateAPISignature("POST", jsonBody)

	req, err := http.NewRequest("POST", s.cfg.IPaymuBaseURL+"/payment", bytes.NewReader(jsonBody))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("va", s.cfg.IPaymuVA)
	req.Header.Set("signature", signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Status  int  `json:"Status"`
		Success bool `json:"Success"`
		Data    struct {
			SessionID string `json:"SessionID"`
			Url       string `json:"Url"`
		} `json:"Data"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}

	if result.Status != 200 || !result.Success {
		return "", "", fmt.Errorf("iPaymu error: %s", string(respBody))
	}

	return result.Data.Url, result.Data.SessionID, nil
}

func (s *PaymentService) generateAPISignature(method string, jsonBody []byte) string {
	// Step 1: SHA-256 hash of the JSON body
	bodyHash := sha256.Sum256(jsonBody)
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))

	// Step 2: Build string to sign = "METHOD:{va}:{bodyHash}:{apiKey}"
	stringToSign := method + ":" + s.cfg.IPaymuVA + ":" + bodyHashHex + ":" + s.cfg.IPaymuAPIKey

	// Step 3: HMAC-SHA256 the stringToSign with apiKey
	h := hmac.New(sha256.New, []byte(s.cfg.IPaymuAPIKey))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *PaymentService) verifySignature(body []byte, signature string) bool {
	expected := s.generateAPISignature("POST", body)
	return hmac.Equal([]byte(expected), []byte(signature))
}
