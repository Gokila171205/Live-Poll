package handlers

import (
	"github.com/gin-gonic/gin"
)

// APIResponse provides a consistent, standardized envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError specifies error details in an APIResponse.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SendSuccess sends a JSON response with success status and attached payload.
func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SendError sends a standardized error response with an HTTP status code, error code, and message.
func SendError(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// SendErrorWithDetails sends a standardized error response including extra context details.
func SendErrorWithDetails(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
