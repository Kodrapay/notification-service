package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/kodra-pay/notification-service/internal/models"
	"github.com/lib/pq"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(dsn string) (*NotificationRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &NotificationRepository{db: db}, nil
}

// DB returns the underlying database connection
func (r *NotificationRepository) DB() *sql.DB {
	return r.db
}

// Create inserts a new notification
func (r *NotificationRepository) Create(ctx context.Context, notif *models.Notification) error {
	templateDataJSON, _ := json.Marshal(notif.TemplateData)
	metadataJSON, _ := json.Marshal(notif.Metadata)

	query := `
		INSERT INTO notifications (
			merchant_id, user_id, type, channel, recipient,
			subject, message, template_name, template_data,
			status, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`
	var merchantID sql.NullInt32
	if notif.MerchantID != nil {
		merchantID.Int32 = int32(*notif.MerchantID)
		merchantID.Valid = true
	}
	var userID sql.NullInt32
	if notif.UserID != nil {
		userID.Int32 = int32(*notif.UserID)
		userID.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx, query,
		merchantID, // *int converted to sql.NullInt32
		userID,     // *int converted to sql.NullInt32
		notif.Type, notif.Channel,
		notif.Recipient, notif.Subject, notif.Message, notif.TemplateName,
		templateDataJSON, notif.Status, metadataJSON,
	).Scan(&notif.ID, &notif.CreatedAt) // notif.ID is int

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id int) (*models.Notification, error) {
	query := `
		SELECT id, merchant_id, user_id, type, channel, recipient,
		       subject, message, template_name, template_data,
		       status, sent_at, delivered_at, error_message,
		       retry_count, metadata, created_at
		FROM notifications
		WHERE id = $1
	`

	var notif models.Notification
	var templateDataJSON, metadataJSON []byte
	var merchantID sql.NullInt32
	var userID sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, id).Scan( // id is int
		&notif.ID,
		&merchantID, // Scan into sql.NullInt32
		&userID,     // Scan into sql.NullInt32
		&notif.Type,
		&notif.Channel, &notif.Recipient, &notif.Subject, &notif.Message,
		&notif.TemplateName, &templateDataJSON, &notif.Status,
		&notif.SentAt, &notif.DeliveredAt, &notif.ErrorMessage,
		&notif.RetryCount, &metadataJSON, &notif.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if len(templateDataJSON) > 0 {
		json.Unmarshal(templateDataJSON, &notif.TemplateData)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &notif.Metadata)
	}

	if merchantID.Valid {
		val := int(merchantID.Int32)
		notif.MerchantID = &val
	}
	if userID.Valid {
		val := int(userID.Int32)
		notif.UserID = &val
	}

	return &notif, nil
}

// UpdateStatus updates the notification status
func (r *NotificationRepository) UpdateStatus(
	ctx context.Context,
	id int, // id changed to int
	status models.NotificationStatus,
	errorMessage *string,
) error {
	query := `
		UPDATE notifications SET
			status = $2,
			error_message = $3,
			sent_at = CASE WHEN $2 = 'sent' OR $2 = 'delivered' THEN NOW() ELSE sent_at END,
			delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END,
			retry_count = retry_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, status, errorMessage) // id is int
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// ListPending retrieves pending notifications for processing
func (r *NotificationRepository) ListPending(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
		SELECT id, merchant_id, user_id, type, channel, recipient,
		       subject, message, template_name, template_data,
		       status, retry_count, metadata, created_at
		FROM notifications
		WHERE status = 'pending'
		  AND retry_count < 3
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows) // Delegate scanning to helper
}

// NotificationPreferencesRepository handles notification preferences
type NotificationPreferencesRepository struct {
	db *sql.DB
}

func NewNotificationPreferencesRepository(db *sql.DB) *NotificationPreferencesRepository {
	return &NotificationPreferencesRepository{db: db}
}

// GetByMerchantID retrieves notification preferences for a merchant
func (r *NotificationPreferencesRepository) GetByMerchantID(
	ctx context.Context,
	merchantID int, // merchantID changed to int
) (*models.NotificationPreferences, error) {
	query := `
		SELECT id, merchant_id, email_enabled, sms_enabled, push_enabled,
		       transaction_notifications, payout_notifications,
		       settlement_notifications, security_notifications,
		       marketing_notifications, email_address, phone_number,
		       created_at, updated_at
		FROM notification_preferences
		WHERE merchant_id = $1
	`

	var prefs models.NotificationPreferences
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan( // merchantID is int
		&prefs.ID, &prefs.MerchantID, // int, int
		&prefs.EmailEnabled, &prefs.SMSEnabled,
		&prefs.PushEnabled, &prefs.TransactionNotifications,
		&prefs.PayoutNotifications, &prefs.SettlementNotifications,
		&prefs.SecurityNotifications, &prefs.MarketingNotifications,
		&prefs.EmailAddress, &prefs.PhoneNumber,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences if none exist
		return r.CreateDefault(ctx, merchantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	return &prefs, nil
}

// CreateDefault creates default notification preferences for a merchant
func (r *NotificationPreferencesRepository) CreateDefault(
	ctx context.Context,
	merchantID int, // merchantID changed to int
) (*models.NotificationPreferences, error) {
	query := `
		INSERT INTO notification_preferences (
			merchant_id, email_enabled, sms_enabled, push_enabled,
			transaction_notifications, payout_notifications,
			settlement_notifications, security_notifications
		) VALUES ($1, TRUE, FALSE, TRUE, TRUE, TRUE, TRUE, TRUE)
		RETURNING id, merchant_id, email_enabled, sms_enabled, push_enabled,
		          transaction_notifications, payout_notifications,
		          settlement_notifications, security_notifications,
		          marketing_notifications, email_address, phone_number,
		          created_at, updated_at
	`

	var prefs models.NotificationPreferences
	err := r.db.QueryRowContext(ctx, query, merchantID).Scan( // merchantID is int
		&prefs.ID, &prefs.MerchantID, // int, int
		&prefs.EmailEnabled, &prefs.SMSEnabled,
		&prefs.PushEnabled, &prefs.TransactionNotifications,
		&prefs.PayoutNotifications, &prefs.SettlementNotifications,
		&prefs.SecurityNotifications, &prefs.MarketingNotifications,
		&prefs.EmailAddress, &prefs.PhoneNumber,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Unique constraint violation, preferences already exist
			return r.GetByMerchantID(ctx, merchantID)
		}
		return nil, fmt.Errorf("failed to create default preferences: %w", err)
	}

	return &prefs, nil
}

// Update updates notification preferences
func (r *NotificationPreferencesRepository) Update(
	ctx context.Context,
	prefs *models.NotificationPreferences,
) error {
	query := `
		UPDATE notification_preferences SET
			email_enabled = $2,
			sms_enabled = $3,
			push_enabled = $4,
			transaction_notifications = $5,
			payout_notifications = $6,
			settlement_notifications = $7,
			security_notifications = $8,
			marketing_notifications = $9,
			email_address = $10,
			phone_number = $11,
			updated_at = NOW()
		WHERE merchant_id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		prefs.MerchantID, prefs.EmailEnabled, prefs.SMSEnabled,
		prefs.PushEnabled, prefs.TransactionNotifications,
		prefs.PayoutNotifications, prefs.SettlementNotifications,
		prefs.SecurityNotifications, prefs.MarketingNotifications,
		prefs.EmailAddress, prefs.PhoneNumber,
	)

	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("preferences not found for merchant: %d", prefs.MerchantID)
	}

	return nil
}

// ListByUserID retrieves all notifications for a specific user
func (r *NotificationRepository) ListByUserID(ctx context.Context, userID int) ([]*models.Notification, error) { // userID changed to int
	query := `
		SELECT id, merchant_id, user_id, type, channel, recipient,
		       subject, message, template_name, template_data,
		       status, sent_at, delivered_at, error_message,
		       retry_count, metadata, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID) // userID is int
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications by user_id: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// ListByMerchantID retrieves all notifications for a specific merchant
func (r *NotificationRepository) ListByMerchantID(ctx context.Context, merchantID int) ([]*models.Notification, error) { // merchantID changed to int
	query := `
		SELECT id, merchant_id, user_id, type, channel, recipient,
		       subject, message, template_name, template_data,
		       status, sent_at, delivered_at, error_message,
		       retry_count, metadata, created_at
		FROM notifications
		WHERE merchant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, merchantID) // merchantID is int
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications by merchant_id: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// scanNotifications is a helper to scan multiple notification rows
func scanNotifications(rows *sql.Rows) ([]*models.Notification, error) {
	var notifications []*models.Notification
	for rows.Next() {
		var notif models.Notification
		var templateDataJSON, metadataJSON []byte
		var merchantID sql.NullInt32
		var userID sql.NullInt32

		err := rows.Scan(
			&notif.ID,          // int
			&merchantID,       // Scan into sql.NullInt32
			&userID,           // Scan into sql.NullInt32
			&notif.Type,
			&notif.Channel, &notif.Recipient, &notif.Subject, &notif.Message,
			&notif.TemplateName, &templateDataJSON, &notif.Status,
			&notif.SentAt, &notif.DeliveredAt, &notif.ErrorMessage,
			&notif.RetryCount, &metadataJSON, &notif.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(templateDataJSON) > 0 {
			json.Unmarshal(templateDataJSON, &notif.TemplateData)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &notif.Metadata)
		}

		// Handle nullable merchantID and userID
		if merchantID.Valid {
			val := int(merchantID.Int32)
			notif.MerchantID = &val
		} else {
			notif.MerchantID = nil
		}
		if userID.Valid {
			val := int(userID.Int32)
			notif.UserID = &val
		} else {
			notif.UserID = nil
		}

		notifications = append(notifications, &notif)
	}

	return notifications, nil
}
