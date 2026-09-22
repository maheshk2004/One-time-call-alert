package routes

import (
	"net/http"
	"time"

	"lead-followup-system/internal/controllers"
	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	AuthService     *services.AuthService
	AuthController  *controllers.AuthController
	LeadController  *controllers.LeadController
	CallController  *controllers.CallController
	TransController *controllers.TranscriptController
	AIController    *controllers.AIController
	FollowUpCtrl    *controllers.FollowUpController
	NotifController *controllers.NotificationController
	DashController  *controllers.DashboardController
	AdminController *controllers.AdminController
	WebhookCtrl     *controllers.WebhookController
}

func SetupRouter(deps *RouterDependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "lead-followup-system"})
	})

	api := r.Group("/api")
	{
		// Public Auth
		api.POST("/auth/login", deps.AuthController.Login)

		// Public Webhook Ingestion
		api.POST("/webhooks/leads/:source", deps.WebhookCtrl.IngestLeadWebhook)

		// Authenticated Routes
		authRequired := api.Group("")
		authRequired.Use(middleware.AuthMiddleware(deps.AuthService))
		{
			// Auth Profile & Users
			authRequired.GET("/auth/me", deps.AuthController.GetCurrentUser)
			authRequired.GET("/users", middleware.RequireRole(models.RoleAdmin, models.RoleManager), deps.AuthController.ListUsers)
			authRequired.POST("/auth/register", middleware.RequireRole(models.RoleAdmin), deps.AuthController.Register)

			// Leads Management & BDA Call Queue
			authRequired.GET("/leads", deps.LeadController.ListLeads)
			authRequired.GET("/bda/queue", deps.LeadController.GetBdaCallQueue)
			authRequired.POST("/leads", deps.LeadController.CreateLead)
			authRequired.GET("/leads/one-call", deps.LeadController.ListOneCallFollowUps)
			authRequired.GET("/leads/do-not-contact", deps.LeadController.ListDoNotContact)
			authRequired.GET("/leads/:id", deps.LeadController.GetLead)
			authRequired.PUT("/leads/:id/assign", middleware.RequireRole(models.RoleAdmin, models.RoleManager), deps.LeadController.AssignLead)
			authRequired.PUT("/leads/:id/status", deps.LeadController.UpdateStatus)
			authRequired.PUT("/leads/:id/do-not-call", deps.LeadController.SetDoNotCall)

			// Call Tracking
			authRequired.POST("/leads/:id/calls", deps.CallController.RecordCall)
			authRequired.GET("/leads/:id/calls", deps.CallController.GetCallsForLead)
			authRequired.GET("/calls/:id", deps.CallController.GetCall)

			// Transcripts
			authRequired.POST("/calls/:id/transcript", deps.TransController.SaveTranscript)
			authRequired.GET("/calls/:id/transcript", deps.TransController.GetTranscript)

			// AI Analysis & Review Queue
			authRequired.POST("/calls/:id/analyze", deps.AIController.AnalyzeCall)
			authRequired.GET("/calls/:id/analysis", deps.AIController.GetAnalysis)
			authRequired.GET("/ai/reviews", deps.AIController.GetPendingReviews)
			authRequired.POST("/ai/reviews/:id/confirm", deps.AIController.ConfirmReview)
			authRequired.POST("/ai/reviews/:id/correct", deps.AIController.CorrectReview)

			// Follow-ups
			authRequired.GET("/follow-ups", deps.FollowUpCtrl.ListFollowUps)
			authRequired.POST("/leads/:id/follow-ups", deps.FollowUpCtrl.ScheduleFollowUp)
			authRequired.POST("/follow-ups/:id/complete", deps.FollowUpCtrl.CompleteFollowUp)
			authRequired.POST("/follow-ups/detect", deps.FollowUpCtrl.TriggerDetection)

			// Notifications Center
			authRequired.GET("/notifications", deps.NotifController.GetNotifications)
			authRequired.PUT("/notifications/:id/read", deps.NotifController.MarkRead)
			authRequired.PUT("/notifications/read-all", deps.NotifController.MarkAllRead)

			// Dashboard & Analytics
			authRequired.GET("/dashboard/summary", deps.DashController.GetSummary)

			// Admin & Settings
			authRequired.GET("/admin/settings", deps.AdminController.GetSettings)
			authRequired.PUT("/admin/settings", middleware.RequireRole(models.RoleAdmin), deps.AdminController.UpdateSettings)
			authRequired.GET("/admin/audit-logs", middleware.RequireRole(models.RoleAdmin, models.RoleManager), deps.AdminController.GetAuditLogs)
		}
	}

	return r
}
