package controllers

import (
	"strconv"
	"time"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	settingsRepo *repositories.SettingsRepository
	auditService *services.AuditService
}

func NewAdminController(settingsRepo *repositories.SettingsRepository, auditService *services.AuditService) *AdminController {
	return &AdminController{
		settingsRepo: settingsRepo,
		auditService: auditService,
	}
}

func (ctrl *AdminController) GetSettings(c *gin.Context) {
	settings, err := ctrl.settingsRepo.GetSettings(c.Request.Context())
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve settings")
		return
	}
	utils.SendSuccess(c, settings)
}

type UpdateSettingsRequest struct {
	FollowUpThresholdHours int      `json:"followUpThresholdHours"`
	AIConfidenceThreshold  float64  `json:"aiConfidenceThreshold"`
	AutoClassifyEnabled    bool     `json:"autoClassifyEnabled"`
	HumanReviewRequiredDNC bool     `json:"humanReviewRequiredDNC"`
	EscalationHours        int      `json:"escalationHours"`
	SupportedLanguages     []string `json:"supportedLanguages"`
	DuplicateCheckFields   []string `json:"duplicateCheckFields"`
}

func (ctrl *AdminController) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	actor := middleware.GetCurrentUser(c)
	current, _ := ctrl.settingsRepo.GetSettings(c.Request.Context())
	if current == nil {
		current = &models.SystemSettings{}
	}

	current.FollowUpThresholdHours = req.FollowUpThresholdHours
	current.AIConfidenceThreshold = req.AIConfidenceThreshold
	current.AutoClassifyEnabled = req.AutoClassifyEnabled
	current.HumanReviewRequiredDNC = req.HumanReviewRequiredDNC
	current.EscalationHours = req.EscalationHours
	if len(req.SupportedLanguages) > 0 {
		current.SupportedLanguages = req.SupportedLanguages
	}
	if len(req.DuplicateCheckFields) > 0 {
		current.DuplicateCheckFields = req.DuplicateCheckFields
	}
	current.UpdatedBy = &actor.ID
	current.UpdatedAt = time.Now()

	if err := ctrl.settingsRepo.UpdateSettings(c.Request.Context(), current); err != nil {
		utils.SendInternalError(c, "Failed to save settings")
		return
	}

	ctrl.auditService.Log(c.Request.Context(), actor.ID, actor.Name, "SETTINGS_UPDATED", "settings", current.ID.Hex(), nil, map[string]interface{}{
		"followUpThresholdHours": req.FollowUpThresholdHours,
		"aiConfidenceThreshold":  req.AIConfidenceThreshold,
	}, "Admin updated system parameters", nil)

	utils.SendSuccess(c, current)
}

func (ctrl *AdminController) GetAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	logs, err := ctrl.auditService.GetRecentLogs(c.Request.Context(), limit)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve audit logs")
		return
	}

	utils.SendSuccess(c, logs)
}
