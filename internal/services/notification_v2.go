package services

import (
	"context"
	"fmt"
	"log"

	"github.com/kodra-pay/notification-service/internal/models"
	"github.com/kodra-pay/notification-service/internal/repositories"
)

type NotificationServiceV2 struct {
	repo      *repositories.NotificationRepository
	prefsRepo *repositories.NotificationPreferencesRepository
}

func NewNotificationServiceV2(
	repo *repositories.NotificationRepository,
	prefsRepo *repositories.NotificationPreferencesRepository,
) *NotificationServiceV2 {
	return &NotificationServiceV2{
		repo:      repo,
		prefsRepo: prefsRepo,
	}
}

// Send creates and sends a notification
func (s *NotificationServiceV2) Send(ctx context.Context, notif *models.Notification) error {
	// Get merchant's notification preferences
	if notif.MerchantID != nil {
		prefs, err := s.prefsRepo.GetByMerchantID(ctx, *notif.MerchantID)
		if err != nil {
			log.Printf("Failed to get notification preferences: %v", err)
			// Continue anyway - use defaults
		} else {
			// Check if notification should be sent based on preferences
			if !prefs.ShouldSend(notif.Type, notif.Channel) {
				return fmt.Errorf("notification disabled by merchant preferences")
			}

			// Use preference contact info if not specified
			if notif.Recipient == "" {
				if notif.Type == models.TypeEmail && prefs.EmailAddress != nil {
					notif.Recipient = *prefs.EmailAddress
				} else if notif.Type == models.TypeSMS && prefs.PhoneNumber != nil {
					notif.Recipient = *prefs.PhoneNumber
				}
			}
		}
	}

	// Validate recipient
	if notif.Recipient == "" {
		return fmt.Errorf("recipient is required")
	}

	// Create notification in database
	notif.Status = models.StatusPending
	if err := s.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification based on type
	var err error
	switch notif.Type {
	case models.TypeEmail:
		err = s.sendEmail(ctx, notif)
	case models.TypeSMS:
		err = s.sendSMS(ctx, notif)
	case models.TypePush:
		err = s.sendPush(ctx, notif)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notif.Type)
	}

	// Update status based on result
	if err != nil {
		errMsg := err.Error()
		s.repo.UpdateStatus(ctx, notif.ID, models.StatusFailed, &errMsg)
		return err
	}

	s.repo.UpdateStatus(ctx, notif.ID, models.StatusSent, nil)
	return nil
}

// sendEmail sends an email notification
func (s *NotificationServiceV2) sendEmail(ctx context.Context, notif *models.Notification) error {
	// TODO: Integrate with email service (SendGrid, AWS SES, etc.)
	log.Printf("Sending email to %s: %s", notif.Recipient, notif.Message)

	// For now, just log the email
	// In production, you would call an email provider API here
	log.Printf("[EMAIL] To: %s", notif.Recipient)
	if notif.Subject != nil {
		log.Printf("[EMAIL] Subject: %s", *notif.Subject)
	}
	log.Printf("[EMAIL] Message: %s", notif.Message)

	return nil
}

// sendSMS sends an SMS notification
func (s *NotificationServiceV2) sendSMS(ctx context.Context, notif *models.Notification) error {
	// TODO: Integrate with SMS service (Twilio, Termii, etc.)
	log.Printf("Sending SMS to %s: %s", notif.Recipient, notif.Message)

	// For now, just log the SMS
	// In production, you would call an SMS provider API here
	log.Printf("[SMS] To: %s", notif.Recipient)
	log.Printf("[SMS] Message: %s", notif.Message)

	return nil
}

// sendPush sends a push notification
func (s *NotificationServiceV2) sendPush(ctx context.Context, notif *models.Notification) error {
	// TODO: Integrate with push notification service (Firebase, OneSignal, etc.)
	log.Printf("Sending push notification to %s: %s", notif.Recipient, notif.Message)

	// For now, just log the push notification
	// In production, you would call a push notification provider API here
	log.Printf("[PUSH] To: %s", notif.Recipient)
	if notif.Subject != nil {
		log.Printf("[PUSH] Title: %s", *notif.Subject)
	}
	log.Printf("[PUSH] Message: %s", notif.Message)

	return nil
}

// SendTransactionNotification sends a transaction-related notification
func (s *NotificationServiceV2) SendTransactionNotification(
	ctx context.Context,
	merchantID string,
	recipient string,
	amount int64,
	currency string,
	status string,
) error {
	subject := "Transaction Notification"
	message := fmt.Sprintf(
		"Transaction of %s %d has been %s",
		currency, amount/100, status,
	)

	notif := &models.Notification{
		MerchantID: &merchantID,
		Type:       models.TypeEmail,
		Channel:    models.ChannelTransaction,
		Recipient:  recipient,
		Subject:    &subject,
		Message:    message,
	}

	return s.Send(ctx, notif)
}

// SendPayoutNotification sends a payout-related notification
func (s *NotificationServiceV2) SendPayoutNotification(
	ctx context.Context,
	merchantID string,
	recipient string,
	amount int64,
	currency string,
	status string,
) error {
	subject := "Payout Notification"
	message := fmt.Sprintf(
		"Payout of %s %d has been %s",
		currency, amount/100, status,
	)

	notif := &models.Notification{
		MerchantID: &merchantID,
		Type:       models.TypeEmail,
		Channel:    models.ChannelPayout,
		Recipient:  recipient,
		Subject:    &subject,
		Message:    message,
	}

	return s.Send(ctx, notif)
}
