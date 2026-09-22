package controllers

import (
	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallController struct {
	callService *services.CallService
}

func NewCallController(callService *services.CallService) *CallController {
	return &CallController{
		callService: callService,
	}
}

func (ctrl *CallController) RecordCall(c *gin.Context) {
	leadIDStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(leadIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	var input services.RecordCallInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}
	input.LeadID = leadID

	actor := middleware.GetCurrentUser(c)
	call, err := ctrl.callService.RecordCall(c.Request.Context(), input, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, call)
}

func (ctrl *CallController) GetCallsForLead(c *gin.Context) {
	leadIDStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(leadIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	calls, err := ctrl.callService.GetCallsForLead(c.Request.Context(), leadID)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve calls")
		return
	}

	utils.SendSuccess(c, calls)
}

func (ctrl *CallController) GetCall(c *gin.Context) {
	callIDStr := c.Param("id")
	callID, err := primitive.ObjectIDFromHex(callIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid call ID")
		return
	}

	call, err := ctrl.callService.GetCallByID(c.Request.Context(), callID)
	if err != nil {
		utils.SendNotFound(c, "Call record not found")
		return
	}

	utils.SendSuccess(c, call)
}
