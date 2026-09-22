package controllers

import (
	"strconv"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationController struct {
	notifService *services.NotificationService
}

func NewNotificationController(notifService *services.NotificationService) *NotificationController {
	return &NotificationController{
		notifService: notifService,
	}
}

func (ctrl *NotificationController) GetNotifications(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	unreadOnly := c.Query("unreadOnly") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifs, err := ctrl.notifService.GetUserNotifications(c.Request.Context(), actor.ID, unreadOnly, limit)
	if err != nil {
		utils.SendInternalError(c, "Failed to load notifications")
		return
	}

	unreadCount, _ := ctrl.notifService.GetUnreadCount(c.Request.Context(), actor.ID)

	utils.SendSuccess(c, gin.H{
		"notifications": notifs,
		"unreadCount":   unreadCount,
	})
}

func (ctrl *NotificationController) MarkRead(c *gin.Context) {
	idStr := c.Param("id")
	notifID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendBadRequest(c, "Invalid notification ID")
		return
	}

	actor := middleware.GetCurrentUser(c)
	if err := ctrl.notifService.MarkRead(c.Request.Context(), notifID, actor.ID); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendSuccess(c, gin.H{"message": "Notification marked as read"})
}

func (ctrl *NotificationController) MarkAllRead(c *gin.Context) {
	actor := middleware.GetCurrentUser(c)
	if err := ctrl.notifService.MarkAllRead(c.Request.Context(), actor.ID); err != nil {
		utils.SendInternalError(c, "Failed to mark notifications read")
		return
	}

	utils.SendSuccess(c, gin.H{"message": "All notifications marked as read"})
}
