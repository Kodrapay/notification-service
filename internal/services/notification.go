package services

import (
	"context"
	"fmt"
	"time"

	"github.com/kodra-pay/notification-service/internal/dto"
	"github.com/kodra-pay/notification-service/internal/models"
	"github.com/kodra-pay/notification-service/internal/repositories"
)

type NotificationService struct {
	repo *repositories.NotificationRepository
}

func NewNotificationService(repo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Send(ctx context.Context, req dto.NotificationRequest) (dto.NotificationResponse, error) {
	notif := &models.Notification{
		Channel:   models.NotificationChannel(req.Channel),
		Recipient: req.To,
		Subject:   &req.Subject,
		Message:   req.Body,
		Status:    models.StatusPending,
	}
	if err := s.repo.Create(ctx, notif); err != nil {
		return dto.NotificationResponse{}, err
	}
	return dto.NotificationResponse{ID: notif.ID, Status: string(notif.Status)}, nil
}

func (s *NotificationService) Get(ctx context.Context, id string) (dto.NotificationResponse, error) {
	notif, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.NotificationResponse{}, fmt.Errorf("notification not found")
	}
	sentAt := ""
	if notif.SentAt != nil {
		sentAt = notif.SentAt.Format(time.RFC3339)
	}
	return dto.NotificationResponse{
		ID:        notif.ID,
		Status:    string(notif.Status),
        SentAt:    sentAt,
	}, nil
}
