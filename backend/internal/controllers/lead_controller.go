package controllers

import (
	"strconv"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadController struct {
	leadService *services.LeadService
}

func NewLeadController(leadService *services.LeadService) *LeadController {
	return &LeadController{
		leadService: leadService,
	}
}

func (ctrl *LeadController) CreateLead(c *gin.Context) {
	var input services.CreateLeadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	actor := middleware.GetCurrentUser(c)
	lead, err := ctrl.leadService.CreateLead(c.Request.Context(), input, actor)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, lead)
}

func (ctrl *LeadController) GetLead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	lead, err := ctrl.leadService.GetLeadByID(c.Request.Context(), id)
	if err != nil {
		utils.SendNotFound(c, "Lead not found")
		return
	}

	utils.SendSuccess(c, lead)
}

func (ctrl *LeadController) ListLeads(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var assignedTo *primitive.ObjectID
	// If salesperson, only list their leads unless specified
	if actor.Role == models.RoleSales {
		assignedTo = &actor.ID
	} else if assignedToStr := c.Query("assignedTo"); assignedToStr != "" {
		if aid, err := primitive.ObjectIDFromHex(assignedToStr); err == nil {
			assignedTo = &aid
		}
	}

	filter := repositories.LeadFilter{
		AssignedTo:     assignedTo,
		Status:         models.LeadStatus(c.Query("status")),
		Priority:       models.Priority(c.Query("priority")),
		Source:         c.Query("source"),
		Campaign:       c.Query("campaign"),
		FollowUpStatus: models.FollowUpStatus(c.Query("followUpStatus")),
		Search:         c.Query("search"),
	}

	if dncStr := c.Query("doNotCall"); dncStr != "" {
		dnc := dncStr == "true"
		filter.DoNotCall = &dnc
	}

	leads, total, err := ctrl.leadService.ListLeads(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve leads")
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, utils.PaginatedData{
		Items:      leads,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (ctrl *LeadController) ListOneCallFollowUps(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}

	leads, total, err := ctrl.leadService.ListOneCallFollowUps(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve one-call follow-ups")
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, utils.PaginatedData{
		Items:      leads,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (ctrl *LeadController) ListDoNotContact(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}

	leads, total, err := ctrl.leadService.ListDoNotCallLeads(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.SendInternalError(c, "Failed to retrieve do not contact leads")
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, utils.PaginatedData{
		Items:      leads,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (ctrl *LeadController) GetBdaCallQueue(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	category := c.DefaultQuery("category", "all")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}

	var assignedTo *primitive.ObjectID
	if actor.Role == models.RoleSales && c.Query("allReps") != "true" {
		assignedTo = &actor.ID
	}

	leads, total, counts, err := ctrl.leadService.GetBdaCallQueue(c.Request.Context(), assignedTo, category, search, page, pageSize)
	if err != nil {
		utils.SendInternalError(c, err.Error())
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	utils.SendSuccess(c, gin.H{
		"items":          leads,
		"totalCount":     total,
		"page":           page,
		"pageSize":       pageSize,
		"totalPages":     totalPages,
		"categoryCounts": counts,
	})
}


type AssignRequest struct {
	SalespersonID string `json:"salespersonId" binding:"required"`
}

func (ctrl *LeadController) AssignLead(c *gin.Context) {
	idStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	salespersonID, err := primitive.ObjectIDFromHex(req.SalespersonID)
	if err != nil {
		utils.SendBadRequest(c, "Invalid salesperson ID")
		return
	}

	actor := middleware.GetCurrentUser(c)
	if err := ctrl.leadService.AssignLead(c.Request.Context(), leadID, salespersonID, actor); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{"message": "Lead successfully assigned"})
}

type UpdateStatusRequest struct {
	Status models.LeadStatus `json:"status" binding:"required"`
	Reason string            `json:"reason"`
}

func (ctrl *LeadController) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	actor := middleware.GetCurrentUser(c)
	if err := ctrl.leadService.UpdateStatus(c.Request.Context(), leadID, req.Status, actor, req.Reason); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{"message": "Lead status updated"})
}

type SetDoNotCallRequest struct {
	DoNotCall bool   `json:"doNotCall"`
	Reason    string `json:"reason" binding:"required"`
}

func (ctrl *LeadController) SetDoNotCall(c *gin.Context) {
	idStr := c.Param("id")
	leadID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid lead ID")
		return
	}

	var req SetDoNotCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "Reason is required to update Do Not Call status")
		return
	}

	actor := middleware.GetCurrentUser(c)
	if err := ctrl.leadService.SetDoNotCall(c.Request.Context(), leadID, req.DoNotCall, req.Reason, actor); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{"message": "Do Not Call status updated"})
}
