package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/services"
)

// AuthHandler handles HTTP requests for user authentication.
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler initializes a new AuthHandler.
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Signup handles new user registration.
// POST /api/auth/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid signup request format. Valid email and password (minimum 8 characters) required.")
		return
	}

	res, err := h.authService.Signup(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyInUse) {
			SendError(c, http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists.")
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user. Please try again later.")
		return
	}

	SendSuccess(c, http.StatusCreated, res)
}

// Login handles user authentication and JWT generation.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid login request format. Email and password required.")
		return
	}

	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			SendError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password.")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Authentication failed. Please try again later.")
		return
	}

	SendSuccess(c, http.StatusOK, res)
}

// GetCurrentUser returns the profile of the authenticated user from the token context.
// GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	val, exists := c.Get(models.ContextKeyUserID)
	if !exists {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	userID, ok := val.(string)
	if !ok || userID == "" {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authentication state")
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		SendError(c, http.StatusNotFound, "USER_NOT_FOUND", "User profile not found")
		return
	}

	SendSuccess(c, http.StatusOK, user)
}
