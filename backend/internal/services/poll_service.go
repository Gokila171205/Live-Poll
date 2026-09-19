package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/redis"
	"live-poll-backend/internal/repository"
)

const (
	minQuestionLength = 3
	maxQuestionLength = 300
	minOptionCount    = 2
	maxOptionCount    = 10
	minOptionLength   = 1
	maxOptionLength   = 100
)

var (
	// ErrPollNotFound indicates the requested poll does not exist.
	ErrPollNotFound = errors.New("poll not found")
	// ErrForbiddenAction indicates the requesting user lacks permission to modify or delete the poll.
	ErrForbiddenAction = errors.New("only the creator of this poll can perform this action")
	// ErrInvalidOption indicates the selected option ID does not belong to the target poll.
	ErrInvalidOption = errors.New("invalid poll option")
	// ErrAlreadyVoted indicates the voter has already voted on this poll.
	ErrAlreadyVoted = repository.ErrAlreadyVoted
	// ErrPollClosed indicates the poll is no longer active or accepting votes.
	ErrPollClosed = errors.New("poll is closed and no longer accepting votes")
)

// PollService provides business logic operations for polls and live vote counting.
type PollService interface {
	CreatePoll(ctx context.Context, creatorID string, req models.CreatePollRequest) (*models.PollResponse, error)
	GetPollByID(ctx context.Context, id string) (*models.PollResponse, error)
	GetMyPolls(ctx context.Context, creatorID string) ([]*models.PollResponse, error)
	DeletePoll(ctx context.Context, id string, creatorID string) error
	SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error
	Vote(ctx context.Context, pollID, optionID, voterID, userID string) (*models.PollResultsResponse, error)
	GetPollResults(ctx context.Context, pollID string) (*models.PollResultsResponse, error)
}

type pollService struct {
	pollRepo    repository.PollRepository
	voteCounter redis.VoteCounter
	publisher   redis.EventPublisher
}

// NewPollService creates a new PollService instance with injected dependencies.
// Notice that PollService does NOT depend on WebSockets; it publishes update events
// strictly through the EventPublisher (Redis Pub/Sub).
func NewPollService(
	pollRepo repository.PollRepository,
	voteCounter redis.VoteCounter,
	publisher redis.EventPublisher,
) PollService {
	return &pollService{
		pollRepo:    pollRepo,
		voteCounter: voteCounter,
		publisher:   publisher,
	}
}

// CreatePoll validates input, constructs option items with unique IDs, and stores the poll.
func (s *pollService) CreatePoll(ctx context.Context, creatorID string, req models.CreatePollRequest) (*models.PollResponse, error) {
	if strings.TrimSpace(creatorID) == "" {
		return nil, fmt.Errorf("%w: creator ID is required", ErrInvalidInput)
	}

	trimmedQuestion := strings.TrimSpace(req.Question)
	if trimmedQuestion == "" {
		return nil, fmt.Errorf("%w: question cannot be empty", ErrInvalidInput)
	}
	if len(trimmedQuestion) < minQuestionLength {
		return nil, fmt.Errorf("%w: question must be at least %d characters long", ErrInvalidInput, minQuestionLength)
	}
	if len(trimmedQuestion) > maxQuestionLength {
		return nil, fmt.Errorf("%w: question exceeds maximum length of %d characters", ErrInvalidInput, maxQuestionLength)
	}

	if len(req.Options) < minOptionCount {
		return nil, fmt.Errorf("%w: poll requires at least %d options", ErrInvalidInput, minOptionCount)
	}
	if len(req.Options) > maxOptionCount {
		return nil, fmt.Errorf("%w: poll cannot have more than %d options", ErrInvalidInput, maxOptionCount)
	}

	seenOptions := make(map[string]struct{}, len(req.Options))
	options := make([]models.PollOption, 0, len(req.Options))
	optionIDs := make([]string, 0, len(req.Options))

	for i, rawOpt := range req.Options {
		trimmedOpt := strings.TrimSpace(rawOpt)
		if trimmedOpt == "" {
			return nil, fmt.Errorf("%w: option %d cannot be empty", ErrInvalidInput, i+1)
		}
		if len(trimmedOpt) < minOptionLength {
			return nil, fmt.Errorf("%w: option %d is too short", ErrInvalidInput, i+1)
		}
		if len(trimmedOpt) > maxOptionLength {
			return nil, fmt.Errorf("%w: option %d exceeds maximum length of %d characters", ErrInvalidInput, i+1, maxOptionLength)
		}

		lower := strings.ToLower(trimmedOpt)
		if _, exists := seenOptions[lower]; exists {
			return nil, fmt.Errorf("%w: duplicate option '%s' detected", ErrInvalidInput, trimmedOpt)
		}
		seenOptions[lower] = struct{}{}

		optID := bson.NewObjectID().Hex()
		options = append(options, models.PollOption{
			ID:        optID,
			Text:      trimmedOpt,
			VoteCount: 0,
		})
		optionIDs = append(optionIDs, optID)
	}

	poll := &models.Poll{
		CreatorID: creatorID,
		Question:  trimmedQuestion,
		Options:   options,
		IsActive:  true,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, fmt.Errorf("failed to persist poll: %w", err)
	}

	// Initialize live option counters in Redis
	if s.voteCounter != nil {
		if err := s.voteCounter.InitPollCounters(ctx, poll.ID.Hex(), optionIDs); err != nil {
			log.Printf("[WARN] Failed to initialize Redis option counters for poll %s: %v", poll.ID.Hex(), err)
		}
	}

	resp := poll.ToResponse()
	return &resp, nil
}

// GetPollByID retrieves a poll by its unique ID.
func (s *pollService) GetPollByID(ctx context.Context, id string) (*models.PollResponse, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil, ErrPollNotFound
	}

	poll, err := s.pollRepo.FindByID(ctx, trimmedID)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("failed to retrieve poll: %w", err)
	}

	resp := poll.ToResponse()
	return &resp, nil
}

// GetMyPolls retrieves all polls created by the given creator ID.
func (s *pollService) GetMyPolls(ctx context.Context, creatorID string) ([]*models.PollResponse, error) {
	trimmedCreatorID := strings.TrimSpace(creatorID)
	if trimmedCreatorID == "" {
		return []*models.PollResponse{}, nil
	}

	polls, err := s.pollRepo.FindByCreatorID(ctx, trimmedCreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user polls: %w", err)
	}

	responses := make([]*models.PollResponse, 0, len(polls))
	for _, p := range polls {
		resp := p.ToResponse()
		responses = append(responses, &resp)
	}

	return responses, nil
}

// DeletePoll validates creator authorization and deletes the specified poll from MongoDB and Redis.
func (s *pollService) DeletePoll(ctx context.Context, id string, creatorID string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return ErrPollNotFound
	}
	trimmedCreatorID := strings.TrimSpace(creatorID)
	if trimmedCreatorID == "" {
		return ErrForbiddenAction
	}

	poll, err := s.pollRepo.FindByID(ctx, trimmedID)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return ErrPollNotFound
		}
		return fmt.Errorf("failed to locate poll for deletion: %w", err)
	}

	if poll.CreatorID != trimmedCreatorID {
		return ErrForbiddenAction
	}

	err = s.pollRepo.Delete(ctx, trimmedID, trimmedCreatorID)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return ErrPollNotFound
		}
		if errors.Is(err, repository.ErrUnauthorizedAction) {
			return ErrForbiddenAction
		}
		return fmt.Errorf("failed to delete poll: %w", err)
	}

	if s.voteCounter != nil {
		optionIDs := make([]string, len(poll.Options))
		for i, o := range poll.Options {
			optionIDs[i] = o.ID
		}
		if err := s.voteCounter.DeletePollCounters(ctx, trimmedID, optionIDs); err != nil {
			log.Printf("[WARN] Failed to delete Redis counters for poll %s: %v", trimmedID, err)
		}
	}

	return nil
}

// SetPollStatus updates the open/closed status of a poll (creator-only).
func (s *pollService) SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return ErrPollNotFound
	}

	err := s.pollRepo.SetPollStatus(ctx, trimmedID, creatorID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return ErrPollNotFound
		}
		if errors.Is(err, repository.ErrUnauthorizedAction) {
			return ErrForbiddenAction
		}
		return fmt.Errorf("failed to update poll status: %w", err)
	}

	results, resErr := s.GetPollResults(ctx, trimmedID)
	if resErr == nil {
		s.broadcastPollUpdate(ctx, results)
	}

	return nil
}

// Vote validates poll activity and option validity, records the vote persistently in MongoDB,
// and atomically increments the Redis live counter with eventual consistency reconciliation.
func (s *pollService) Vote(ctx context.Context, pollID, optionID, voterID, userID string) (*models.PollResultsResponse, error) {
	trimmedPollID := strings.TrimSpace(pollID)
	trimmedOptionID := strings.TrimSpace(optionID)
	trimmedVoterID := strings.TrimSpace(voterID)

	if trimmedPollID == "" {
		return nil, ErrPollNotFound
	}
	if trimmedOptionID == "" {
		return nil, fmt.Errorf("%w: optionId is required", ErrInvalidInput)
	}
	if trimmedVoterID == "" {
		return nil, fmt.Errorf("%w: valid voter identifier is required", ErrInvalidInput)
	}

	// 1. Verify poll exists in persistent storage
	poll, err := s.pollRepo.FindByID(ctx, trimmedPollID)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("failed to find poll: %w", err)
	}

	// 2. Verify the poll is active
	if !poll.IsActive {
		return nil, ErrPollClosed
	}

	// 3. Prevent arbitrary Redis keys: verify option belongs to this poll
	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == trimmedOptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, fmt.Errorf("%w: option %s does not belong to poll %s", ErrInvalidOption, trimmedOptionID, trimmedPollID)
	}

	// 4. Record vote persistently in MongoDB (atomic unique index prevents duplicate votes under race conditions)
	voteRecord := &models.Vote{
		PollID:   trimmedPollID,
		OptionID: trimmedOptionID,
		VoterID:  trimmedVoterID,
		UserID:   userID,
	}
	if err := s.pollRepo.RecordVote(ctx, voteRecord); err != nil {
		if errors.Is(err, repository.ErrAlreadyVoted) {
			return nil, ErrAlreadyVoted
		}
		return nil, fmt.Errorf("failed to record vote in MongoDB: %w", err)
	}

	// 5. Increment option counter atomically in Redis.
	// Consistency Strategy: If Redis fails, we do NOT fail the vote because it is already safely
	// and durably recorded in MongoDB. We log the warning and let GetPollResults reconcile Redis.
	if s.voteCounter != nil {
		_, rErr := s.voteCounter.IncrementOption(ctx, trimmedPollID, trimmedOptionID)
		if rErr != nil {
			log.Printf("[WARN] Redis atomic increment failed for poll %s, option %s: %v. Eventual consistency will reconcile via MongoDB.", trimmedPollID, trimmedOptionID, rErr)
		}
	}

	// 6. Return updated live results
	results, err := s.GetPollResults(ctx, trimmedPollID)
	if err != nil {
		return nil, err
	}

	// 7. Publish and broadcast live update event to Redis Pub/Sub and WebSocket clients
	s.broadcastPollUpdate(ctx, results)

	return results, nil
}

func (s *pollService) broadcastPollUpdate(ctx context.Context, results *models.PollResultsResponse) {
	if results == nil {
		return
	}
	event := results.ToEvent()
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WARN] Failed to marshal poll update event: %v", err)
		return
	}

	// Publish to Redis Pub/Sub channel. The backend Redis subscriber will receive this
	// and broadcast to all connected WebSocket clients for that poll room.
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, results.PollID, payload); pubErr != nil {
			log.Printf("[WARN] Redis Pub/Sub publish failed for poll %s: %v", results.PollID, pubErr)
		}
	}
}

// GetPollResults retrieves live vote counts from Redis with automatic reconciliation from MongoDB.
func (s *pollService) GetPollResults(ctx context.Context, pollID string) (*models.PollResultsResponse, error) {
	trimmedPollID := strings.TrimSpace(pollID)
	if trimmedPollID == "" {
		return nil, ErrPollNotFound
	}

	poll, err := s.pollRepo.FindByID(ctx, trimmedPollID)
	if err != nil {
		if errors.Is(err, repository.ErrPollNotFound) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("failed to find poll for results: %w", err)
	}

	optionIDs := make([]string, len(poll.Options))
	persistentCounts := make(map[string]int64, len(poll.Options))
	for i, opt := range poll.Options {
		optionIDs[i] = opt.ID
		persistentCounts[opt.ID] = opt.VoteCount
	}

	// Read live counts from Redis
	var liveCounts map[string]int64
	if s.voteCounter != nil {
		counts, err := s.voteCounter.GetPollResults(ctx, trimmedPollID, optionIDs)
		if err != nil {
			log.Printf("[WARN] Redis unavailable for live results on poll %s: %v. Reconciling from MongoDB.", trimmedPollID, err)
		} else {
			liveCounts = counts
		}
	}

	// Build results and reconcile any missing, expired, or lagging Redis keys
	var totalVotes int64
	resultOptions := make([]models.VoteResultOption, len(poll.Options))

	for i, opt := range poll.Options {
		var count int64
		if liveCounts != nil {
			redisCount, exists := liveCounts[opt.ID]
			if exists && redisCount >= 0 {
				count = redisCount
			} else {
				// Missing or expired key in Redis: restore from MongoDB persistent count
				count = persistentCounts[opt.ID]
				if s.voteCounter != nil {
					_ = s.voteCounter.SetOptionCount(ctx, trimmedPollID, opt.ID, count)
				}
			}
		} else {
			count = persistentCounts[opt.ID]
		}

		resultOptions[i] = models.VoteResultOption{
			ID:        opt.ID,
			Text:      opt.Text,
			VoteCount: count,
		}
		totalVotes += count
	}

	// Calculate percentages
	for i := range resultOptions {
		if totalVotes > 0 {
			pct := (float64(resultOptions[i].VoteCount) / float64(totalVotes)) * 100
			resultOptions[i].Percentage = math.Round(pct*10) / 10
		} else {
			resultOptions[i].Percentage = 0.0
		}
	}

	return &models.PollResultsResponse{
		PollID:     poll.ID.Hex(),
		Question:   poll.Question,
		IsActive:   poll.IsActive,
		TotalVotes: totalVotes,
		Options:    resultOptions,
	}, nil
}
