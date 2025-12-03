package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/notification-service/internal/handlers"
	"github.com/kodra-pay/notification-service/internal/services"
)

func Register(app *fiber.App, service string) {
	health := handlers.NewHealthHandler(service)
	health.Register(app)

	svc := services.NewNotificationService()
	h := handlers.NewNotificationHandler(svc)
	api := app.Group("/notifications")
	api.Post("/", h.Send)
	api.Get("/:id", h.Get)
}
