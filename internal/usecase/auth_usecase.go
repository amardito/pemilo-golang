package usecase

import (
	"fmt"
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/amardito/pemilo-golang/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	MaxFailedAttempts = 3
	LockoutDuration   = 5 * time.Minute
	TokenExpiry       = 24 * time.Hour
)

type AuthUsecase struct {
	adminRepo        domain.AdminRepository
	loginAttemptRepo domain.LoginAttemptRepository
	jwtSecret        string
	encryptionKey    string
}

func NewAuthUsecase(
	adminRepo domain.AdminRepository,
	loginAttemptRepo domain.LoginAttemptRepository,
	jwtSecret string,
	encryptionKey string,
) *AuthUsecase {
	return &AuthUsecase{
		adminRepo:        adminRepo,
		loginAttemptRepo: loginAttemptRepo,
		jwtSecret:        jwtSecret,
		encryptionKey:    encryptionKey,
	}
}

// Login authenticates admin with rate limiting (3 failures = 5 min lockout)
func (u *AuthUsecase) Login(username, encryptedPassword string) (*domain.Admin, string, time.Time, error) {
	// Check rate limiting
	since := time.Now().Add(-LockoutDuration)
	failedAttempts, err := u.loginAttemptRepo.GetRecentFailedAttempts(username, since)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	if failedAttempts >= MaxFailedAttempts {
		// Check if lockout period has passed
		lastAttempt, err := u.loginAttemptRepo.GetLastAttempt(username)
		if err != nil {
			return nil, "", time.Time{}, err
		}

		if lastAttempt != nil && time.Since(lastAttempt.AttemptAt) < LockoutDuration {
			remainingTime := LockoutDuration - time.Since(lastAttempt.AttemptAt)
			return nil, "", time.Time{}, fmt.Errorf("%w: try again in %v", domain.ErrRateLimitExceeded, remainingTime.Round(time.Second))
		}
	}

	// Get admin by username
	admin, err := u.adminRepo.GetByUsername(username)
	if err != nil {
		u.recordLoginAttempt(username, false)
		if err == domain.ErrAdminNotFound {
			return nil, "", time.Time{}, domain.ErrInvalidCredentials
		}
		return nil, "", time.Time{}, err
	}

	// Check if admin is active
	if !admin.IsActive {
		u.recordLoginAttempt(username, false)
		return nil, "", time.Time{}, domain.ErrAdminInactive
	}

	// Verify password (bcrypt hash comparison)
	if err := utils.VerifyPassword(admin.Password, encryptedPassword); err != nil {
		u.recordLoginAttempt(username, false)
		return nil, "", time.Time{}, domain.ErrInvalidCredentials
	}

	// Successful login - record attempt
	u.recordLoginAttempt(username, true)

	// Generate JWT token
	token, expiresAt, err := u.generateToken(admin)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return admin, token, expiresAt, nil
}

// generateToken creates JWT token for authenticated admin
func (u *AuthUsecase) generateToken(admin *domain.Admin) (string, time.Time, error) {
	expiresAt := time.Now().Add(TokenExpiry)

	claims := jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// recordLoginAttempt logs login attempt for rate limiting
func (u *AuthUsecase) recordLoginAttempt(identifier string, success bool) {
	attempt := &domain.LoginAttempt{
		ID:         uuid.New().String(),
		Identifier: identifier,
		AttemptAt:  time.Now(),
		Success:    success,
	}
	_ = u.loginAttemptRepo.RecordAttempt(attempt)
}

// CleanupOldAttempts removes old login attempts (run periodically)
func (u *AuthUsecase) CleanupOldAttempts() error {
	before := time.Now().Add(-24 * time.Hour)
	return u.loginAttemptRepo.CleanupOldAttempts(before)
}
