package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/kodra-pay/notification-service/internal/dto"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService { return &NotificationService{} }

func (s *NotificationService) Send(_ context.Context, req dto.NotificationRequest) dto.NotificationResponse {
	return dto.NotificationResponse{ID: "notif_" + uuid.NewString(), Status: "queued"}
}

func (s *NotificationService) Get(_ context.Context, id string) dto.NotificationResponse {
	return dto.NotificationResponse{ID: id, Status: "queued"}
}
