package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lead-followup-system/internal/ai"
	"lead-followup-system/internal/config"
	"lead-followup-system/internal/controllers"
	"lead-followup-system/internal/database"
	"lead-followup-system/internal/repositories"
	"lead-followup-system/internal/routes"
	"lead-followup-system/internal/services"
)

func main() {
	log.Println("Starting Lead Follow-Up Intelligence System API...")

	cfg := config.LoadConfig()

	// MongoDB
	mongoDB, err := database.ConnectMongo(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Redis
	redisService, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Repositories
	userRepo := repositories.NewUserRepository(mongoDB.Database)
	leadRepo := repositories.NewLeadRepository(mongoDB.Database)
	callRepo := repositories.NewCallRepository(mongoDB.Database)
	transRepo := repositories.NewTranscriptRepository(mongoDB.Database)
	aiRepo := repositories.NewAIRepository(mongoDB.Database)
	followUpRepo := repositories.NewFollowUpRepository(mongoDB.Database)
	notifRepo := repositories.NewNotificationRepository(mongoDB.Database)
	auditRepo := repositories.NewAuditRepository(mongoDB.Database)
	settingsRepo := repositories.NewSettingsRepository(mongoDB.Database)

	// Ensure system settings exists
	_, _ = settingsRepo.GetSettings(context.Background())

	// AI Provider
	aiProvider := ai.NewLLMProvider(cfg.OpenAIAPIKey)

	// Services
	authService := services.NewAuthService(userRepo, cfg)
	auditService := services.NewAuditService(auditRepo)
	notifService := services.NewNotificationService(notifRepo)
	emailService := services.NewEmailService(cfg, redisService.Client)
	leadService := services.NewLeadService(leadRepo, userRepo, followUpRepo, auditService, notifService)
	callService := services.NewCallService(callRepo, leadRepo, userRepo, followUpRepo, auditService)
	transService := services.NewTranscriptService(transRepo, callRepo, auditService)
	aiService := services.NewAIService(aiRepo, callRepo, leadRepo, transRepo, userRepo, settingsRepo, followUpRepo, auditService, notifService, aiProvider)
	followUpService := services.NewFollowUpService(followUpRepo, leadRepo, userRepo, settingsRepo, notifService, emailService, auditService)
	dashService := services.NewDashboardService(leadRepo, callRepo, followUpRepo, aiRepo, userRepo)

	// Controllers
	authCtrl := controllers.NewAuthController(authService, userRepo)
	leadCtrl := controllers.NewLeadController(leadService)
	callCtrl := controllers.NewCallController(callService)
	transCtrl := controllers.NewTranscriptController(transService)
	aiCtrl := controllers.NewAIController(aiService)
	followUpCtrl := controllers.NewFollowUpController(followUpService)
	notifCtrl := controllers.NewNotificationController(notifService)
	dashCtrl := controllers.NewDashboardController(dashService)
	adminCtrl := controllers.NewAdminController(settingsRepo, auditService)
	webhookCtrl := controllers.NewWebhookController(leadService)

	router := routes.SetupRouter(&routes.RouterDependencies{
		AuthService:     authService,
		AuthController:  authCtrl,
		LeadController:  leadCtrl,
		CallController:  callCtrl,
		TransController: transCtrl,
		AIController:    aiCtrl,
		FollowUpCtrl:    followUpCtrl,
		NotifController: notifCtrl,
		DashController:  dashCtrl,
		AdminController: adminCtrl,
		WebhookCtrl:     webhookCtrl,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully.")
}
