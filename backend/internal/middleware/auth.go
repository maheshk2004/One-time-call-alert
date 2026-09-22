package middleware

import (
	"strings"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/services"
	"lead-followup-system/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserKey   = "currentUser"
	ContextClaimsKey = "userClaims"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendUnauthorized(c, "Authorization token is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendUnauthorized(c, "Invalid Authorization header format. Expected 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			utils.SendUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		user, err := authService.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil || user == nil || !user.IsActive {
			utils.SendUnauthorized(c, "User account not found or inactive")
			c.Abort()
			return
		}

		c.Set(ContextUserKey, user)
		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userObj, exists := c.Get(ContextUserKey)
		if !exists {
			utils.SendUnauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		user, ok := userObj.(*models.User)
		if !ok {
			utils.SendUnauthorized(c, "Invalid user context")
			c.Abort()
			return
		}

		hasRole := false
		for _, r := range roles {
			if user.Role == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.SendForbidden(c, "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *models.User {
	if val, ok := c.Get(ContextUserKey); ok {
		if user, ok := val.(*models.User); ok {
			return user
		}
	}
	return nil
}
