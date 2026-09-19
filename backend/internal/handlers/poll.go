package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/services"
)

var safeIdentifierRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{1,64}$`)

func isValidObjectID(id string) bool {
	id = strings.TrimSpace(id)
	if len(id) != 24 {
		return false
	}
	_, err := bson.ObjectIDFromHex(id)
	return err == nil
}

func sanitizeIdentifier(val string) string {
	val = strings.TrimSpace(val)
	if len(val) == 0 || len(val) > 64 {
		return ""
	}
	if !safeIdentifierRegex.MatchString(val) {
		return ""
	}
	return val
}

// PollHandler manages HTTP endpoints related to poll operations.
type PollHandler struct {
	pollService services.PollService
}

// NewPollHandler creates a new instance of PollHandler with injected service.
func NewPollHandler(pollService services.PollService) *PollHandler {
	return &PollHandler{pollService: pollService}
}

// CreatePoll handles creation of a new poll by an authenticated user.
// POST /api/polls
func (h *PollHandler) CreatePoll(c *gin.Context) {
	val, exists := c.Get(models.ContextKeyUserID)
	userID, ok := val.(string)
	if !exists || !ok || userID == "" {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required to create a poll")
		return
	}

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "INVALID_INPUT", "Question and at least 2 options are required in JSON body")
		return
	}

	poll, err := h.pollService.CreatePoll(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create poll. Please try again later.")
		return
	}

	SendSuccess(c, http.StatusCreated, poll)
}

// GetPoll retrieves an individual poll by its ID.
// GET /api/polls/:id
func (h *PollHandler) GetPoll(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(id) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	poll, err := h.pollService.GetPollByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve poll")
		return
	}

	SendSuccess(c, http.StatusOK, poll)
}

// GetMyPolls retrieves all polls created by the authenticated user.
// GET /api/my/polls
func (h *PollHandler) GetMyPolls(c *gin.Context) {
	val, exists := c.Get(models.ContextKeyUserID)
	userID, ok := val.(string)
	if !exists || !ok || userID == "" {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	polls, err := h.pollService.GetMyPolls(c.Request.Context(), userID)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve your polls")
		return
	}

	SendSuccess(c, http.StatusOK, polls)
}

// DeletePoll deletes a poll if the authenticated user is the creator.
// DELETE /api/polls/:id
func (h *PollHandler) DeletePoll(c *gin.Context) {
	val, exists := c.Get(models.ContextKeyUserID)
	userID, ok := val.(string)
	if !exists || !ok || userID == "" {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required to delete a poll")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(id) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	err := h.pollService.DeletePoll(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		if errors.Is(err, services.ErrForbiddenAction) {
			SendError(c, http.StatusForbidden, "FORBIDDEN", "You are not authorized to delete this poll")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete poll")
		return
	}

	SendSuccess(c, http.StatusOK, gin.H{
		"message": "Poll deleted successfully",
		"pollId":  id,
	})
}

// SetPollStatus updates the active state of a poll (creator-only).
// PATCH /api/polls/:id/status
func (h *PollHandler) SetPollStatus(c *gin.Context) {
	val, exists := c.Get(models.ContextKeyUserID)
	userID, ok := val.(string)
	if !exists || !ok || userID == "" {
		SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(id) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	var req models.SetPollStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid status payload. Expected JSON: { \"isActive\": boolean }")
		return
	}

	err := h.pollService.SetPollStatus(c.Request.Context(), id, userID, req.IsActive)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		if errors.Is(err, services.ErrForbiddenAction) {
			SendError(c, http.StatusForbidden, "FORBIDDEN", "You are not authorized to change status of this poll")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update poll status")
		return
	}

	SendSuccess(c, http.StatusOK, gin.H{
		"pollId":   id,
		"isActive": req.IsActive,
	})
}

// GetPollResults returns the live option vote counts from Redis for a given poll.
// GET /api/polls/:id/results
func (h *PollHandler) GetPollResults(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(id) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	results, err := h.pollService.GetPollResults(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve live poll results")
		return
	}

	SendSuccess(c, http.StatusOK, results)
}

// Vote handles casting an audience vote on a poll option.
// The audience does NOT need creator authentication to vote.
// POST /api/polls/:id/vote
func (h *PollHandler) Vote(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if !isValidObjectID(id) {
		SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
		return
	}

	var req models.CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OptionID) == "" {
		SendError(c, http.StatusBadRequest, "INVALID_INPUT", "A valid optionId is required in JSON body: { \"optionId\": \"...\" }")
		return
	}

	cleanOptionID := strings.TrimSpace(req.OptionID)
	if !safeIdentifierRegex.MatchString(cleanOptionID) {
		SendError(c, http.StatusBadRequest, "INVALID_OPTION", "Invalid option ID format")
		return
	}

	// Resolve voter identifier safely
	voterID, userID := resolveVoterIdentifier(c)

	results, err := h.pollService.Vote(c.Request.Context(), id, cleanOptionID, voterID, userID)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			SendError(c, http.StatusNotFound, "POLL_NOT_FOUND", "The requested poll does not exist")
			return
		}
		if errors.Is(err, services.ErrPollClosed) {
			SendError(c, http.StatusBadRequest, "POLL_CLOSED", "This poll is closed and no longer accepting votes")
			return
		}
		if errors.Is(err, services.ErrInvalidOption) {
			SendError(c, http.StatusBadRequest, "INVALID_OPTION", err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, services.ErrAlreadyVoted) {
			SendError(c, http.StatusConflict, "ALREADY_VOTED", "You have already cast a vote in this poll")
			return
		}
		SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record vote")
		return
	}

	SendSuccess(c, http.StatusOK, results)
}

// resolveVoterIdentifier determines a safe voter ID without trusting arbitrary unvalidated payloads.
func resolveVoterIdentifier(c *gin.Context) (voterID string, userID string) {
	// 1. Authenticated User (if verified token was extracted by Auth or OptionalAuth middleware)
	if val, exists := c.Get(models.ContextKeyUserID); exists {
		if uid, ok := val.(string); ok && strings.TrimSpace(uid) != "" {
			cleanUID := strings.TrimSpace(uid)
			return "user:" + cleanUID, cleanUID
		}
	}

	// 2. Audience Header (X-Voter-ID or X-Session-ID)
	if h := sanitizeIdentifier(c.GetHeader("X-Voter-ID")); h != "" {
		return "header:" + h, ""
	}
	if h := sanitizeIdentifier(c.GetHeader("X-Session-ID")); h != "" {
		return "header:" + h, ""
	}

	// 3. Client Cookie
	if cookie, err := c.Cookie("livepoll_voter_id"); err == nil {
		if cleanCookie := sanitizeIdentifier(cookie); cleanCookie != "" {
			return "cookie:" + cleanCookie, ""
		}
	}

	// 4. Generate cryptographically safe session ID and attach HTTP cookie
	newID := bson.NewObjectID().Hex()
	c.SetCookie("livepoll_voter_id", newID, 86400*30, "/", "", false, true)
	c.Header("X-Voter-ID", newID)
	return "cookie:" + newID, ""
}
