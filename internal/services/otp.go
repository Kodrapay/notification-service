package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kodra-pay/notification-service/internal/models"
	"github.com/kodra-pay/notification-service/internal/repositories"
)

type OTPService struct {
	otpRepo       *repositories.OTPRepository
	notifService  *NotificationServiceV2
}

func NewOTPService(
	otpRepo *repositories.OTPRepository,
	notifService *NotificationServiceV2,
) *OTPService {
	return &OTPService{
		otpRepo:      otpRepo,
		notifService: notifService,
	}
}

// Generate creates and sends a new OTP
func (s *OTPService) Generate(ctx context.Context, req *models.CreateOTPRequest) (*models.OTP, error) {
	// Set defaults
	if req.ExpiryMinutes == 0 {
		req.ExpiryMinutes = 10 // Default 10 minutes
	}
	if req.MaxAttempts == 0 {
		req.MaxAttempts = 3 // Default 3 attempts
	}

	// Generate OTP code
	code, err := models.GenerateCode(6) // 6-digit code
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	// Create OTP record
	otp := &models.OTP{
		MerchantID:     req.MerchantID,
		UserID:         req.UserID,
		Purpose:        req.Purpose,
		Code:           code,
		Recipient:      req.Recipient,
		DeliveryMethod: req.DeliveryMethod,
		ExpiresAt:      time.Now().Add(time.Duration(req.ExpiryMinutes) * time.Minute),
		Attempts:       0,
		MaxAttempts:    req.MaxAttempts,
		ReferenceID:    req.ReferenceID,
		Metadata:       req.Metadata,
	}

	// Save to database
	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Send OTP via notification
	if err := s.sendOTP(ctx, otp); err != nil {
		return nil, fmt.Errorf("failed to send OTP: %w", err)
	}

	// Don't return the actual code in the response for security
	otpResponse := *otp
	otpResponse.Code = "******" // Mask the code
	return &otpResponse, nil
}

// Verify validates an OTP code
func (s *OTPService) Verify(ctx context.Context, req *models.VerifyOTPRequest) (*models.OTP, error) {
	// Get OTP by code
	otp, err := s.otpRepo.GetByCode(ctx, req.MerchantID, req.Purpose, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	// If reference ID is provided, validate it matches
	if req.ReferenceID != nil && otp.ReferenceID != nil {
		if *req.ReferenceID != *otp.ReferenceID {
			return nil, fmt.Errorf("OTP reference mismatch")
		}
	}

	// Attempt verification
	err = otp.Verify(req.Code)

	// Update attempts count
	s.otpRepo.UpdateAttempts(ctx, otp.ID, otp.Attempts)

	if err != nil {
		return nil, err
	}

	// Mark as verified
	if err := s.otpRepo.MarkAsVerified(ctx, otp.ID); err != nil {
		return nil, fmt.Errorf("failed to mark OTP as verified: %w", err)
	}

	return otp, nil
}

// Resend generates and sends a new OTP for the same purpose and reference
func (s *OTPService) Resend(ctx context.Context, req *models.CreateOTPRequest) (*models.OTP, error) {
	// Invalidate existing OTPs for this reference
	if req.ReferenceID != nil {
		s.otpRepo.InvalidateByReferenceID(ctx, req.MerchantID, req.Purpose, *req.ReferenceID)
	}

	// Generate new OTP
	return s.Generate(ctx, req)
}

// sendOTP sends the OTP code via the specified delivery method
func (s *OTPService) sendOTP(ctx context.Context, otp *models.OTP) error {
	var message string
	var notifType models.NotificationType

	// Build OTP message based on purpose
	switch otp.Purpose {
	case models.PurposePayout:
		message = fmt.Sprintf("Your KodraPay payout verification code is: %s. Valid for %d minutes. Do not share this code with anyone.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	case models.PurposeWithdrawal:
		message = fmt.Sprintf("Your KodraPay withdrawal verification code is: %s. Valid for %d minutes. Do not share this code with anyone.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	case models.PurposeSettingsChange:
		message = fmt.Sprintf("Your KodraPay settings change verification code is: %s. Valid for %d minutes.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	case models.PurposeLogin:
		message = fmt.Sprintf("Your KodraPay login verification code is: %s. Valid for %d minutes.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	case models.Purpose2FA:
		message = fmt.Sprintf("Your KodraPay 2FA code is: %s. Valid for %d minutes.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	default:
		message = fmt.Sprintf("Your KodraPay verification code is: %s. Valid for %d minutes.",
			otp.Code, int(otp.ExpiresAt.Sub(otp.CreatedAt).Minutes()))
	}

	// Determine notification type based on delivery method
	switch otp.DeliveryMethod {
	case models.DeliveryEmail:
		notifType = models.TypeEmail
	case models.DeliverySMS:
		notifType = models.TypeSMS
	default:
		return fmt.Errorf("unsupported delivery method: %s", otp.DeliveryMethod)
	}

	// Send notification
	subject := "KodraPay Verification Code"
	notif := &models.Notification{
		MerchantID: &otp.MerchantID,
		UserID:     otp.UserID,
		Type:       notifType,
		Channel:    models.ChannelSecurity,
		Recipient:  otp.Recipient,
		Subject:    &subject,
		Message:    message,
	}

	return s.notifService.Send(ctx, notif)
}

// CleanupExpired removes expired OTPs from the database
func (s *OTPService) CleanupExpired(ctx context.Context) (int64, error) {
	// Delete OTPs expired more than 24 hours ago
	return s.otpRepo.CleanupExpired(ctx, 24*time.Hour)
}
