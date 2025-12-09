package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/notification-service/internal/handlers"
	"github.com/kodra-pay/notification-service/internal/config"
	"github.com/kodra-pay/notification-service/internal/repositories"
	"github.com/kodra-pay/notification-service/internal/services"
)

func Register(app *fiber.App, serviceName string) {
	health := handlers.NewHealthHandler(serviceName)
	health.Register(app)

	cfg := config.Load(serviceName, "7014")

	// connect DB
	repo, err := repositories.NewNotificationRepository(cfg.PostgresDSN)
	if err != nil {
		panic(err)
	}

	// Notification service
	notifSvc := services.NewNotificationService(repo)
	notifHandler := handlers.NewNotificationHandler(notifSvc)

	// OTP service
	otpRepo := repositories.NewOTPRepository(repo.DB())
	prefsRepo := repositories.NewNotificationPreferencesRepository(repo.DB())
	notifSvcV2 := services.NewNotificationServiceV2(repo, prefsRepo)
	otpSvc := services.NewOTPService(otpRepo, notifSvcV2)
	otpHandler := handlers.NewOTPHandler(otpSvc)

	// Notification routes
	app.Post("/notifications", notifHandler.Send)
	app.Get("/notifications/:id", notifHandler.Get)
	app.Get("/notifications/user/:userID", notifHandler.ListByUserID)
	app.Get("/notifications/merchant/:merchantID", notifHandler.ListByMerchantID)

	// OTP routes
	app.Post("/otp/generate", otpHandler.Generate)
	app.Post("/otp/verify", otpHandler.Verify)
	app.Post("/otp/resend", otpHandler.Resend)
}
