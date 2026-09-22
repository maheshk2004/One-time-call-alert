package controllers

import (
	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TranscriptController struct {
	transcriptService *services.TranscriptService
}

func NewTranscriptController(transcriptService *services.TranscriptService) *TranscriptController {
	return &TranscriptController{
		transcriptService: transcriptService,
	}
}

func (ctrl *TranscriptController) SaveTranscript(c *gin.Context) {
	callIDStr := c.Param("id")
	callID, err := primitive.ObjectIDFromHex(callIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid call ID")
		return
	}

	var input services.SaveTranscriptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}
	input.CallAttemptID = callID

	actor := middleware.GetCurrentUser(c)
	transcript, err := ctrl.transcriptService.SaveTranscript(c.Request.Context(), input, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, transcript)
}

func (ctrl *TranscriptController) GetTranscript(c *gin.Context) {
	callIDStr := c.Param("id")
	callID, err := primitive.ObjectIDFromHex(callIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid call ID")
		return
	}

	transcript, err := ctrl.transcriptService.GetByCallAttemptID(c.Request.Context(), callID)
	if err != nil {
		utils.SendNotFound(c, "Transcript not found for this call")
		return
	}

	utils.SendSuccess(c, transcript)
}
