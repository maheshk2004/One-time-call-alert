package controllers

import (
	"net/http"

	"lead-followup-system/internal/middleware"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
	userRepo    *repositories.UserRepository
}

func NewAuthController(authService *services.AuthService, userRepo *repositories.UserRepository) *AuthController {
	return &AuthController{
		authService: authService,
		userRepo:    userRepo,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, "Email and password are required")
		return
	}

	token, user, err := ctrl.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err.Error(), "INVALID_CREDENTIALS")
		return
	}

	utils.SendSuccess(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"teamId":   user.TeamID,
			"isActive": user.IsActive,
		},
	})
}

func (ctrl *AuthController) GetCurrentUser(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.SendUnauthorized(c, "Not authenticated")
		return
	}

	utils.SendSuccess(c, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"teamId":   user.TeamID,
		"phone":    user.Phone,
		"isActive": user.IsActive,
	})
}

type RegisterRequest struct {
	Name     string      `json:"name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     models.Role `json:"role" binding:"required"`
	TeamID   string      `json:"teamId"`
	Phone    string      `json:"phone"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	user, err := ctrl.authService.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.Role, req.TeamID, req.Phone)
	if err != nil {
		utils.SendBadRequest(c, err.Error())
		return
	}

	utils.SendCreated(c, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"isActive": user.IsActive,
	})
}

func (ctrl *AuthController) ListUsers(c *gin.Context) {
	roleStr := c.Query("role")
	var role models.Role
	if roleStr != "" {
		role = models.Role(roleStr)
	}

	users, err := ctrl.userRepo.FindAll(c.Request.Context(), role)
	if err != nil {
		utils.SendInternalError(c, "Failed to load users")
		return
	}

	var response []gin.H
	for _, u := range users {
		response = append(response, gin.H{
			"id":       u.ID,
			"name":     u.Name,
			"email":    u.Email,
			"role":     u.Role,
			"teamId":   u.TeamID,
			"isActive": u.IsActive,
		})
	}

	utils.SendSuccess(c, response)
}
