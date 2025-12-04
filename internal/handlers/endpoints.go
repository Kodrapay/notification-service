package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/notification-service/internal/dto"
	"github.com/kodra-pay/notification-service/internal/services"
)

type NotificationHandler struct {
	svc *services.NotificationService
}

func NewNotificationHandler(svc *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Send(c *fiber.Ctx) error {
	var req dto.NotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.svc.Send(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp)
}

func (h *NotificationHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	resp, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(resp)
}
