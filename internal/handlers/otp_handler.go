package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/notification-service/internal/models"
	"github.com/kodra-pay/notification-service/internal/services"
)

type OTPHandler struct {
	otpService *services.OTPService
}

func NewOTPHandler(otpService *services.OTPService) *OTPHandler {
	return &OTPHandler{otpService: otpService}
}

// Generate creates and sends a new OTP
func (h *OTPHandler) Generate(c *fiber.Ctx) error {
	var req models.CreateOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.MerchantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	if req.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "purpose is required",
		})
	}
	if req.Recipient == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "recipient is required",
		})
	}
	if req.DeliveryMethod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "delivery_method is required (email or sms)",
		})
	}

	otp, err := h.otpService.Generate(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "OTP sent successfully",
		"otp":     otp,
	})
}

// Verify validates an OTP code
func (h *OTPHandler) Verify(c *fiber.Ctx) error {
	var req models.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.MerchantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	if req.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "purpose is required",
		})
	}
	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "code is required",
		})
	}

	otp, err := h.otpService.Verify(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":    err.Error(),
			"verified": false,
		})
	}

	return c.JSON(fiber.Map{
		"message":  "OTP verified successfully",
		"verified": true,
		"otp":      otp,
	})
}

// Resend generates and sends a new OTP
func (h *OTPHandler) Resend(c *fiber.Ctx) error {
	var req models.CreateOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.MerchantID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "merchant_id is required",
		})
	}
	if req.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "purpose is required",
		})
	}
	if req.Recipient == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "recipient is required",
		})
	}
	if req.DeliveryMethod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "delivery_method is required",
		})
	}

	otp, err := h.otpService.Resend(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "OTP resent successfully",
		"otp":     otp,
	})
}
