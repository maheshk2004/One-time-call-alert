package controllers

import (
	"strconv"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AIController struct {
	aiService *services.AIService
}

func NewAIController(aiService *services.AIService) *AIController {
	return &AIController{
		aiService: aiService,
	}
}

func (ctrl *AIController) AnalyzeCall(c *gin.Context) {
	callIDStr := c.Param("id")
	callID, err := primitive.ObjectIDFromHex(callIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid call ID")
		return
	}

	actor := middleware.GetCurrentUser(c)
	analysis, err := ctrl.aiService.AnalyzeCallConversation(c.Request.Context(), callID, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, analysis)
}

func (ctrl *AIController) GetAnalysis(c *gin.Context) {
	callIDStr := c.Param("id")
	callID, err := primitive.ObjectIDFromHex(callIDStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid call ID")
		return
	}

	analysis, err := ctrl.aiService.GetAnalysisByCallID(c.Request.Context(), callID)
	if err != nil {
		utils.SendNotFound(c, "AI analysis not found for this call")
		return
	}

	utils.SendSuccess(c, analysis)
}

func (ctrl *AIController) GetPendingReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}

	analyses, total, err := ctrl.aiService.GetPendingReviews(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve pending reviews")
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, utils.PaginatedData{
		Items:      analyses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (ctrl *AIController) ConfirmReview(c *gin.Context) {
	idStr := c.Param("id")
	analysisID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid analysis ID")
		return
	}

	actor := middleware.GetCurrentUser(c)
	analysis, err := ctrl.aiService.ConfirmAnalysis(c.Request.Context(), analysisID, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, analysis)
}

type CorrectReviewRequest struct {
	Intent models.AIIntent `json:"intent" binding:"required"`
	Reason string          `json:"reason" binding:"required"`
}

func (ctrl *AIController) CorrectReview(c *gin.Context) {
	idStr := c.Param("id")
	analysisID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid analysis ID")
		return
	}

	var req CorrectReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "Intent and reason are required to correct classification")
		return
	}

	actor := middleware.GetCurrentUser(c)
	analysis, err := ctrl.aiService.OverrideAnalysis(c.Request.Context(), analysisID, req.Intent, req.Reason, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, analysis)
}
