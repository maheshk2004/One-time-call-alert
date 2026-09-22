package controllers

import (
	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService *services.DashboardService
}

func NewDashboardController(dashboardService *services.DashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

func (ctrl *DashboardController) GetSummary(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	metrics, err := ctrl.dashboardService.GetSummaryMetrics(c.Request.Context(), actor)
	if err != nil {
		utils.SendInternalError(c, "Failed to load dashboard metrics")
		return
	}

	utils.SendSuccess(c, metrics)
}
