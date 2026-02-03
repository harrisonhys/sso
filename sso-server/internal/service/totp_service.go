package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"image/png"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidTOTPCode   = errors.New("invalid TOTP code")
	ErrTwoFactorNotSetup = errors.New("two-factor authentication not setup")
)

// TOTPService handles TOTP operations
type TOTPService struct {
	userRepo *repository.UserRepository
	issuer   string
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(userRepo *repository.UserRepository, issuer string) *TOTPService {
	return &TOTPService{
		userRepo: userRepo,
		issuer:   issuer,
	}
}

// SetupTOTP generates a new TOTP secret for a user
func (s *TOTPService) SetupTOTP(ctx context.Context, userID string) (*otp.Key, []string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: user.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, nil, err
	}

	// Generate backup codes
	backupCodes := s.generateBackupCodes(10)

	// Hash backup codes for storage
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hash, _ := bcrypt.GenerateFromPassword([]byte(code), 10)
		hashedCodes[i] = string(hash)
	}

	// Store in database (will be marked as enabled after verification)
	twoFA := &models.TwoFactorAuth{
		ID:                   uuid.New().String(),
		UserID:               userID,
		SecretEncrypted:      key.Secret(),
		Enabled:              false,
		BackupCodesEncrypted: string(hashedCodes[0]), // Store as JSON in production
	}

	// Check if 2FA record exists
	if user.TwoFactorAuth != nil {
		twoFA.ID = user.TwoFactorAuth.ID
	}

	return key, backupCodes, nil
}

// VerifyAndEnableTOTP verifies the TOTP code and enables 2FA
func (s *TOTPService) VerifyAndEnableTOTP(ctx context.Context, userID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorAuth == nil {
		return ErrTwoFactorNotSetup
	}

	// Verify code
	valid := totp.Validate(code, user.TwoFactorAuth.SecretEncrypted)
	if !valid {
		return ErrInvalidTOTPCode
	}

	// Enable 2FA
	user.TwoFactorAuth.Enabled = true
	now := time.Now()
	user.TwoFactorAuth.EnabledAt = &now

	return s.userRepo.Update(ctx, user)
}

// ValidateTOTP validates a TOTP code
func (s *TOTPService) ValidateTOTP(ctx context.Context, userID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorAuth == nil || !user.TwoFactorAuth.Enabled {
		return ErrTwoFactorNotSetup
	}

	// Try TOTP code
	valid := totp.Validate(code, user.TwoFactorAuth.SecretEncrypted)
	if valid {
		return nil
	}

	// TODO: Try backup codes validation here

	return ErrInvalidTOTPCode
}

// DisableTOTP disables 2FA for a user
func (s *TOTPService) DisableTOTP(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorAuth != nil {
		user.TwoFactorAuth.Enabled = false
		user.TwoFactorAuth.EnabledAt = nil
		return s.userRepo.Update(ctx, user)
	}

	return nil
}

// GenerateQRCode generates a QR code image for the TOTP secret
func (s *TOTPService) GenerateQRCode(key *otp.Key, writer io.Writer) error {
	img, err := key.Image(200, 200)
	if err != nil {
		return err
	}

	return png.Encode(writer, img)
}

// generateBackupCodes generates random backup codes
func (s *TOTPService) generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		bytes := make([]byte, 6)
		rand.Read(bytes)
		codes[i] = base64.StdEncoding.EncodeToString(bytes)[:8]
	}
	return codes
}

// VerifyCode verifies a TOTP code for an already-enabled 2FA user (used during login)
func (s *TOTPService) VerifyCode(ctx context.Context, userID, code string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.TwoFactorAuth == nil || !user.TwoFactorAuth.Enabled {
		return false, ErrTwoFactorNotSetup
	}

	// Verify code using TOTP library
	valid := totp.Validate(code, user.TwoFactorAuth.SecretEncrypted)
	return valid, nil
}
