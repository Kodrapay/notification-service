package models

import (
	"time"
)

type NotificationType string
type NotificationChannel string
type NotificationStatus string

const (
	TypeEmail NotificationType = "email"
	TypeSMS   NotificationType = "sms"
	TypePush  NotificationType = "push"

	ChannelTransaction NotificationChannel = "transaction"
	ChannelPayout      NotificationChannel = "payout"
	ChannelSettlement  NotificationChannel = "settlement"
	ChannelSecurity    NotificationChannel = "security"
	ChannelSystem      NotificationChannel = "system"

	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusDelivered NotificationStatus = "delivered"
)

type Notification struct {
	ID           string                 `json:"id" db:"id"`
	MerchantID   *string                `json:"merchant_id,omitempty" db:"merchant_id"`
	UserID       *string                `json:"user_id,omitempty" db:"user_id"`
	Type         NotificationType       `json:"type" db:"type"`
	Channel      NotificationChannel    `json:"channel" db:"channel"`
	Recipient    string                 `json:"recipient" db:"recipient"`
	Subject      *string                `json:"subject,omitempty" db:"subject"`
	Message      string                 `json:"message" db:"message"`
	TemplateName *string                `json:"template_name,omitempty" db:"template_name"`
	TemplateData map[string]interface{} `json:"template_data,omitempty" db:"template_data"`
	Status       NotificationStatus     `json:"status" db:"status"`
	SentAt       *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty" db:"delivered_at"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

type NotificationPreferences struct {
	ID                       string     `json:"id" db:"id"`
	MerchantID               string     `json:"merchant_id" db:"merchant_id"`
	EmailEnabled             bool       `json:"email_enabled" db:"email_enabled"`
	SMSEnabled               bool       `json:"sms_enabled" db:"sms_enabled"`
	PushEnabled              bool       `json:"push_enabled" db:"push_enabled"`
	TransactionNotifications bool       `json:"transaction_notifications" db:"transaction_notifications"`
	PayoutNotifications      bool       `json:"payout_notifications" db:"payout_notifications"`
	SettlementNotifications  bool       `json:"settlement_notifications" db:"settlement_notifications"`
	SecurityNotifications    bool       `json:"security_notifications" db:"security_notifications"`
	MarketingNotifications   bool       `json:"marketing_notifications" db:"marketing_notifications"`
	EmailAddress             *string    `json:"email_address,omitempty" db:"email_address"`
	PhoneNumber              *string    `json:"phone_number,omitempty" db:"phone_number"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
}

// ShouldSend determines if a notification should be sent based on preferences
func (np *NotificationPreferences) ShouldSend(notifType NotificationType, channel NotificationChannel) bool {
	// Security notifications are always sent
	if channel == ChannelSecurity && np.SecurityNotifications {
		return true
	}

	// Check channel-specific preference
	var channelEnabled bool
	switch channel {
	case ChannelTransaction:
		channelEnabled = np.TransactionNotifications
	case ChannelPayout:
		channelEnabled = np.PayoutNotifications
	case ChannelSettlement:
		channelEnabled = np.SettlementNotifications
	case ChannelSecurity:
		channelEnabled = np.SecurityNotifications
	case ChannelSystem:
		channelEnabled = true // System notifications are always sent
	default:
		return false
	}

	if !channelEnabled {
		return false
	}

	// Check notification type preference
	switch notifType {
	case TypeEmail:
		return np.EmailEnabled
	case TypeSMS:
		return np.SMSEnabled
	case TypePush:
		return np.PushEnabled
	default:
		return false
	}
}
