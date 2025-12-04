package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type OTPPurpose string
type OTPDeliveryMethod string

const (
	PurposePayout          OTPPurpose = "payout"
	PurposeWithdrawal      OTPPurpose = "withdrawal"
	PurposeSettingsChange  OTPPurpose = "settings_change"
	PurposeLogin           OTPPurpose = "login"
	Purpose2FA             OTPPurpose = "2fa"

	DeliveryEmail OTPDeliveryMethod = "email"
	DeliverySMS   OTPDeliveryMethod = "sms"
)

type OTP struct {
	ID             string                 `json:"id" db:"id"`
	MerchantID     string                 `json:"merchant_id" db:"merchant_id"`
	UserID         *string                `json:"user_id,omitempty" db:"user_id"`
	Purpose        OTPPurpose             `json:"purpose" db:"purpose"`
	Code           string                 `json:"code" db:"code"`
	Recipient      string                 `json:"recipient" db:"recipient"`
	DeliveryMethod OTPDeliveryMethod      `json:"delivery_method" db:"delivery_method"`
	ExpiresAt      time.Time              `json:"expires_at" db:"expires_at"`
	VerifiedAt     *time.Time             `json:"verified_at,omitempty" db:"verified_at"`
	Attempts       int                    `json:"attempts" db:"attempts"`
	MaxAttempts    int                    `json:"max_attempts" db:"max_attempts"`
	ReferenceID    *string                `json:"reference_id,omitempty" db:"reference_id"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// GenerateCode generates a random 6-digit OTP code
func GenerateCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	// Generate random digits
	digits := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		digits += fmt.Sprintf("%d", n)
	}

	return digits, nil
}

// IsExpired checks if the OTP has expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsVerified checks if the OTP has been verified
func (o *OTP) IsVerified() bool {
	return o.VerifiedAt != nil
}

// CanAttempt checks if more verification attempts are allowed
func (o *OTP) CanAttempt() bool {
	return o.Attempts < o.MaxAttempts
}

// Verify validates the provided code against the OTP
func (o *OTP) Verify(code string) error {
	if o.IsVerified() {
		return fmt.Errorf("OTP already verified")
	}

	if o.IsExpired() {
		return fmt.Errorf("OTP has expired")
	}

	if !o.CanAttempt() {
		return fmt.Errorf("maximum verification attempts exceeded")
	}

	o.Attempts++

	if o.Code != code {
		return fmt.Errorf("invalid OTP code")
	}

	now := time.Now()
	o.VerifiedAt = &now
	return nil
}

// CreateOTPRequest represents a request to create an OTP
type CreateOTPRequest struct {
	MerchantID      string                 `json:"merchant_id"`
	UserID          *string                `json:"user_id,omitempty"`
	Purpose         OTPPurpose             `json:"purpose"`
	Recipient       string                 `json:"recipient"`
	DeliveryMethod  OTPDeliveryMethod      `json:"delivery_method"`
	ExpiryMinutes   int                    `json:"expiry_minutes"` // Default: 10 minutes
	MaxAttempts     int                    `json:"max_attempts"`   // Default: 3
	ReferenceID     *string                `json:"reference_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// VerifyOTPRequest represents a request to verify an OTP
type VerifyOTPRequest struct {
	MerchantID  string     `json:"merchant_id"`
	Purpose     OTPPurpose `json:"purpose"`
	Code        string     `json:"code"`
	ReferenceID *string    `json:"reference_id,omitempty"`
}
