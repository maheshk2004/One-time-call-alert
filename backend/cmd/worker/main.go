package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lead-followup-system/internal/ai"
	"lead-followup-system/internal/config"
	"lead-followup-system/internal/database"
	"lead-followup-system/internal/repositories"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/workers"

	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Lead Follow-Up Intelligence Background Worker...")

	cfg := config.LoadConfig()

	// MongoDB
	mongoDB, err := database.ConnectMongo(cfg)
	if err != nil {
		log.Fatalf("Worker failed to connect to MongoDB: %v", err)
	}

	// Redis
	redisService, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("Worker failed to connect to Redis: %v", err)
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

	// Services
	auditService := services.NewAuditService(auditRepo)
	notifService := services.NewNotificationService(notifRepo)
	emailService := services.NewEmailService(cfg, redisService.Client)
	aiProvider := ai.NewLLMProvider(cfg.OpenAIAPIKey)
	aiService := services.NewAIService(aiRepo, callRepo, leadRepo, transRepo, userRepo, settingsRepo, followUpRepo, auditService, notifService, aiProvider)
	followUpService := services.NewFollowUpService(followUpRepo, leadRepo, userRepo, settingsRepo, notifService, emailService, auditService)

	// Initialize Worker Server
	asynqServer := asynq.NewServer(
		redisService.RedisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	taskHandler := workers.NewTaskHandler(followUpService, aiService)

	// Start Background Scheduler (Runs One-Call check every 2 minutes)
	scheduler := workers.NewBackgroundScheduler(redisService.AsynqClient)
	scheduler.Start(2 * time.Minute)

	// Graceful shutdown handling
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Stopping scheduler and worker server...")
		scheduler.Stop()
		asynqServer.Stop()
		asynqServer.Shutdown()
		log.Println("Worker cleanly stopped.")
		os.Exit(0)
	}()

	log.Println("Worker server running and waiting for tasks...")
	mux := asynq.NewServeMux()
	mux.HandleFunc(workers.TypeDetectOneCallFollowUps, func(ctx context.Context, t *asynq.Task) error {
		return taskHandler.ProcessTask(ctx, t)
	})
	mux.HandleFunc(workers.TypeAnalyzeCallConversation, func(ctx context.Context, t *asynq.Task) error {
		return taskHandler.ProcessTask(ctx, t)
	})

	if err := asynqServer.Run(mux); err != nil {
		log.Fatalf("Asynq server failed to run: %v", err)
	}
}
