package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kodra-pay/notification-service/internal/models"
)

type OTPRepository struct {
	db *sql.DB
}

func NewOTPRepository(db *sql.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

// Create inserts a new OTP
func (r *OTPRepository) Create(ctx context.Context, otp *models.OTP) error {
	metadataJSON, _ := json.Marshal(otp.Metadata)

	query := `
		INSERT INTO otps (
			merchant_id, user_id, purpose, code, recipient,
			delivery_method, expires_at, attempts, max_attempts,
			reference_id, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		otp.MerchantID, otp.UserID, otp.Purpose, otp.Code,
		otp.Recipient, otp.DeliveryMethod, otp.ExpiresAt,
		otp.Attempts, otp.MaxAttempts, otp.ReferenceID, metadataJSON,
	).Scan(&otp.ID, &otp.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	return nil
}

// GetByCode retrieves an OTP by code for verification
func (r *OTPRepository) GetByCode(
	ctx context.Context,
	merchantID string,
	purpose models.OTPPurpose,
	code string,
) (*models.OTP, error) {
	query := `
		SELECT id, merchant_id, user_id, purpose, code, recipient,
		       delivery_method, expires_at, verified_at, attempts,
		       max_attempts, reference_id, metadata, created_at
		FROM otps
		WHERE merchant_id = $1
		  AND purpose = $2
		  AND code = $3
		  AND verified_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp models.OTP
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, merchantID, purpose, code).Scan(
		&otp.ID, &otp.MerchantID, &otp.UserID, &otp.Purpose,
		&otp.Code, &otp.Recipient, &otp.DeliveryMethod,
		&otp.ExpiresAt, &otp.VerifiedAt, &otp.Attempts,
		&otp.MaxAttempts, &otp.ReferenceID, &metadataJSON, &otp.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("OTP not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &otp.Metadata)
	}

	return &otp, nil
}

// GetByReferenceID retrieves the latest OTP by reference ID
func (r *OTPRepository) GetByReferenceID(
	ctx context.Context,
	merchantID string,
	purpose models.OTPPurpose,
	referenceID string,
) (*models.OTP, error) {
	query := `
		SELECT id, merchant_id, user_id, purpose, code, recipient,
		       delivery_method, expires_at, verified_at, attempts,
		       max_attempts, reference_id, metadata, created_at
		FROM otps
		WHERE merchant_id = $1
		  AND purpose = $2
		  AND reference_id = $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp models.OTP
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, merchantID, purpose, referenceID).Scan(
		&otp.ID, &otp.MerchantID, &otp.UserID, &otp.Purpose,
		&otp.Code, &otp.Recipient, &otp.DeliveryMethod,
		&otp.ExpiresAt, &otp.VerifiedAt, &otp.Attempts,
		&otp.MaxAttempts, &otp.ReferenceID, &metadataJSON, &otp.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("OTP not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &otp.Metadata)
	}

	return &otp, nil
}

// UpdateAttempts increments the verification attempts counter
func (r *OTPRepository) UpdateAttempts(ctx context.Context, id string, attempts int) error {
	query := `
		UPDATE otps SET
			attempts = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, attempts)
	if err != nil {
		return fmt.Errorf("failed to update OTP attempts: %w", err)
	}

	return nil
}

// MarkAsVerified marks an OTP as verified
func (r *OTPRepository) MarkAsVerified(ctx context.Context, id string) error {
	query := `
		UPDATE otps SET
			verified_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as verified: %w", err)
	}

	return nil
}

// CleanupExpired deletes expired OTPs older than specified duration
func (r *OTPRepository) CleanupExpired(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM otps
		WHERE expires_at < NOW() - $1::interval
	`

	result, err := r.db.ExecContext(ctx, query, olderThan.String())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// InvalidateByReferenceID invalidates all OTPs for a reference ID
func (r *OTPRepository) InvalidateByReferenceID(
	ctx context.Context,
	merchantID string,
	purpose models.OTPPurpose,
	referenceID string,
) error {
	query := `
		UPDATE otps SET
			expires_at = NOW() - INTERVAL '1 hour'
		WHERE merchant_id = $1
		  AND purpose = $2
		  AND reference_id = $3
		  AND verified_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, merchantID, purpose, referenceID)
	if err != nil {
		return fmt.Errorf("failed to invalidate OTPs: %w", err)
	}

	return nil
}
