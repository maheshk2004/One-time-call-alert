package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	ErrorCode string      `json:"errorCode,omitempty"`
}

type PaginatedData struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func SendCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func SendError(c *gin.Context, statusCode int, message string, errorCode string) {
	c.JSON(statusCode, APIResponse{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
	})
}

func SendBadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, message, "BAD_REQUEST")
}

func SendUnauthorized(c *gin.Context, message string) {
	SendError(c, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

func SendForbidden(c *gin.Context, message string) {
	SendError(c, http.StatusForbidden, message, "FORBIDDEN")
}

func SendNotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, message, "NOT_FOUND")
}

func SendInternalError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR")
}
