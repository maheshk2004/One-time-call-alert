package controllers

import (
	"strconv"
	"time"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowUpController struct {
	followUpService *services.FollowUpService
}

func NewFollowUpController(followUpService *services.FollowUpService) *FollowUpController {
	return &FollowUpController{
		followUpService: followUpService,
	}
}

type ScheduleRequest struct {
	DueAt time.Time `json:"dueAt" binding:"required"`
	Notes string    `json:"notes"`
}

func (ctrl *FollowUpController) ScheduleFollowUp(c *gin.Context) {
	leadIDStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(leadIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	actor := middleware.GetCurrentUser(c)
	followUp, err := ctrl.followUpService.ScheduleFollowUp(c.Request.Context(), services.ScheduleFollowUpInput{
		LeadID: leadID,
		DueAt:  req.DueAt,
		Notes:  req.Notes,
	}, actor)

	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, followUp)
}

func (ctrl *FollowUpController) CompleteFollowUp(c *gin.Context) {
	idStr := c.Param("id")
	followUpID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid follow-up ID")
		return
	}

	var body struct {
		CallAttemptID string `json:"callAttemptId"`
	}
	_ = c.ShouldBindJSON(&body)

	var callAttemptID *primitive.ObjectID
	if body.CallAttemptID != "" {
		if cid, err := primitive.ObjectIDFromHex(body.CallAttemptID); err == nil {
			callAttemptID = &cid
		}
	}

	actor := middleware.GetCurrentUser(c)
	if err := ctrl.followUpService.CompleteFollowUp(c.Request.Context(), followUpID, callAttemptID, actor); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{"message": "Follow-up marked completed"})
}

func (ctrl *FollowUpController) ListFollowUps(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}

	var salespersonID *primitive.ObjectID
	if actor.Role == models.RoleSales {
		salespersonID = &actor.ID
	} else if spStr := c.Query("salespersonId"); spStr != "" {
		if spID, err := primitive.ObjectIDFromHex(spStr); err == nil {
			salespersonID = &spID
		}
	}

	status := models.FollowUpTaskStatus(c.Query("status"))
	timeRange := c.Query("timeRange") // "today", "overdue", or all

	followUps, total, err := ctrl.followUpService.ListFollowUps(c.Request.Context(), salespersonID, status, timeRange, page, pageSize)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve follow-ups")
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, utils.PaginatedData{
		Items:      followUps,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// TriggerDetection manually triggers the detector job
func (ctrl *FollowUpController) TriggerDetection(c *gin.Context) {
	count, err := ctrl.followUpService.DetectOneCallFollowUps(c.Request.Context())
	if err != nil {
		utils.SendInternalError(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{
		"message":               "One-call follow-up detection executed successfully",
		"leadsFlaggedForAlert": count,
	})
}
