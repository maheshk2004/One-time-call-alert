package controllers

import (
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type WebhookController struct {
	leadService *services.LeadService
}

func NewWebhookController(leadService *services.LeadService) *WebhookController {
	return &WebhookController{
		leadService: leadService,
	}
}

type WebhookLeadPayload struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email"`
	Phone          string `json:"phone" binding:"required"`
	AlternatePhone string `json:"alternatePhone"`
	Campaign       string `json:"campaign"`
	AdName         string `json:"adName"`
	Course         string `json:"course"`
	Location       string `json:"location"`
	Notes          string `json:"notes"`
}

func (ctrl *WebhookController) IngestLeadWebhook(c *gin.Context) {
	source := c.Param("source")
	if source == "" {
		source = "webhook"
	}

	var payload WebhookLeadPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	systemActor := &models.User{
		Name: "Webhook Ingestion (" + source + ")",
		Role: models.RoleAdmin,
	}

	lead, err := ctrl.leadService.CreateLead(c.Request.Context(), services.CreateLeadInput{
		Name:           payload.Name,
		Email:          payload.Email,
		Phone:          payload.Phone,
		AlternatePhone: payload.AlternatePhone,
		Source:         source,
		Campaign:       payload.Campaign,
		AdName:         payload.AdName,
		Course:         payload.Course,
		Location:       payload.Location,
		Priority:       models.PriorityMedium,
		Notes:          payload.Notes,
	}, systemActor)

	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, gin.H{
		"message": "Lead successfully ingested from " + source,
		"leadId":  lead.ID,
	})
}
