package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

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

func (h *NotificationHandler) ListByUserID(c *fiber.Ctx) error {
	userID := c.Params("userID")
	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "user_id must be a valid uuid")
	}
	resp, err := h.svc.ListByUserID(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp)
}

func (h *NotificationHandler) ListByMerchantID(c *fiber.Ctx) error {
	merchantID := c.Params("merchantID")
	if _, err := uuid.Parse(merchantID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id must be a valid uuid")
	}
	resp, err := h.svc.ListByMerchantID(c.Context(), merchantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(resp)
}
