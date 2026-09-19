package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"live-poll-backend/internal/handlers"
	"live-poll-backend/internal/middleware"
	"live-poll-backend/internal/models"
	"live-poll-backend/internal/redis"
	"live-poll-backend/internal/repository"
	"live-poll-backend/internal/services"
	"live-poll-backend/internal/websocket"
)

type memoryPollRepo struct {
	mu    sync.Mutex
	polls map[string]*models.Poll
	votes map[string]bool
}

func newMemoryPollRepo() *memoryPollRepo {
	return &memoryPollRepo{
		polls: make(map[string]*models.Poll),
		votes: make(map[string]bool),
	}
}

func (m *memoryPollRepo) Create(ctx context.Context, poll *models.Poll) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if poll.ID.IsZero() {
		poll.ID = bson.NewObjectID()
	}
	poll.CreatedAt = time.Now().UTC()
	poll.UpdatedAt = time.Now().UTC()
	poll.IsActive = true
	m.polls[poll.ID.Hex()] = poll
	return nil
}

func (m *memoryPollRepo) FindByID(ctx context.Context, id string) (*models.Poll, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.polls[id]
	if !exists {
		return nil, repository.ErrPollNotFound
	}
	return p, nil
}

func (m *memoryPollRepo) FindByCreatorID(ctx context.Context, creatorID string) ([]*models.Poll, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	polls := make([]*models.Poll, 0)
	for _, p := range m.polls {
		if p.CreatorID == creatorID {
			polls = append(polls, p)
		}
	}
	return polls, nil
}

func (m *memoryPollRepo) Delete(ctx context.Context, id string, creatorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.polls[id]
	if !exists {
		return repository.ErrPollNotFound
	}
	if p.CreatorID != creatorID {
		return repository.ErrUnauthorizedAction
	}
	delete(m.polls, id)
	return nil
}

func (m *memoryPollRepo) SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.polls[id]
	if !exists {
		return repository.ErrPollNotFound
	}
	if p.CreatorID != creatorID {
		return repository.ErrUnauthorizedAction
	}
	p.IsActive = isActive
	return nil
}

func (m *memoryPollRepo) EnsureIndexes(ctx context.Context) error {
	return nil
}

func (m *memoryPollRepo) RecordVote(ctx context.Context, vote *models.Vote) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.polls[vote.PollID]
	if !exists {
		return repository.ErrPollNotFound
	}

	key := vote.PollID + ":" + vote.VoterID
	if m.votes[key] {
		return repository.ErrAlreadyVoted
	}

	for i := range p.Options {
		if p.Options[i].ID == vote.OptionID {
			p.Options[i].VoteCount++
			break
		}
	}
	m.votes[key] = true
	return nil
}

func (m *memoryPollRepo) HasVoterVoted(ctx context.Context, pollID, voterID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if voterID == "" {
		return false, nil
	}
	return m.votes[pollID+":"+voterID], nil
}

type memoryVoteCounter struct {
	mu     sync.Mutex
	counts map[string]int64
}

func newMemoryVoteCounter() *memoryVoteCounter {
	return &memoryVoteCounter{counts: make(map[string]int64)}
}

func (c *memoryVoteCounter) InitPollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, optID := range optionIDs {
		c.counts[pollID+":"+optID] = 0
	}
	return nil
}

func (c *memoryVoteCounter) IncrementOption(ctx context.Context, pollID, optionID string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := pollID + ":" + optionID
	c.counts[key]++
	return c.counts[key], nil
}

func (c *memoryVoteCounter) GetOptionCount(ctx context.Context, pollID, optionID string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[pollID+":"+optionID], nil
}

func (c *memoryVoteCounter) GetPollResults(ctx context.Context, pollID string, optionIDs []string) (map[string]int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := make(map[string]int64)
	for _, optID := range optionIDs {
		res[optID] = c.counts[pollID+":"+optID]
	}
	return res, nil
}

func (c *memoryVoteCounter) SetOptionCount(ctx context.Context, pollID, optionID string, count int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[pollID+":"+optionID] = count
	return nil
}

func (c *memoryVoteCounter) DeletePollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, optID := range optionIDs {
		delete(c.counts, pollID+":"+optID)
	}
	return nil
}

func (c *memoryVoteCounter) IsAvailable(ctx context.Context) bool {
	return true
}

func setupPollTestEnvironment() (*gin.Engine, services.AuthService, services.PollService, *memoryVoteCounter) {
	gin.SetMode(gin.TestMode)

	userRepo := newMemoryUserRepo()
	authService := services.NewAuthService(userRepo, "test-poll-secret", 24)
	authHandler := handlers.NewAuthHandler(authService)

	pollRepo := newMemoryPollRepo()
	voteCounter := newMemoryVoteCounter()
	wsHub := websocket.NewHub()
	go wsHub.Run()

	pubsub := redis.NewPubSubService(nil)
	_ = pubsub.Subscribe(context.Background(), func(pollID string, payload []byte) {
		wsHub.BroadcastToPoll(pollID, payload)
	})

	pollService := services.NewPollService(pollRepo, voteCounter, pubsub)
	pollHandler := handlers.NewPollHandler(pollService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, pollService)

	router := gin.New()
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
		}

		api.GET("/polls/:id", pollHandler.GetPoll)
		api.GET("/polls/:id/results", pollHandler.GetPollResults)
		api.POST("/polls/:id/vote", pollHandler.Vote)
		api.GET("/polls/:id/live", wsHandler.ServeWS)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("/polls", pollHandler.CreatePoll)
			protected.GET("/my/polls", pollHandler.GetMyPolls)
			protected.DELETE("/polls/:id", pollHandler.DeletePoll)
			protected.PATCH("/polls/:id/status", pollHandler.SetPollStatus)
		}
	}

	return router, authService, pollService, voteCounter
}

func TestPollEndpoints_FullFlow(t *testing.T) {
	router, authService, _, _ := setupPollTestEnvironment()
	ctx := context.Background()

	// 1. Register Creator
	creatorResp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "creator@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to register creator: %v", err)
	}

	// 2. Create poll
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/polls", bytes.NewBufferString(`{"question":"Favorite editor?","options":["Vim","VSCode","GoLand"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorResp.Token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Data models.PollResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	pollID := createResp.Data.ID
	opt1 := createResp.Data.Options[0].ID
	opt2 := createResp.Data.Options[1].ID

	// 3. Valid audience vote (no creator authentication required)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"`+opt1+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Voter-ID", "voter-alice")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid audience vote, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Duplicate vote by voter-alice -> 409 Conflict
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"`+opt2+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Voter-ID", "voter-alice")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate vote, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Invalid poll ID -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/nonexistent-id/vote", bytes.NewBufferString(`{"optionId":"`+opt1+`"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for invalid poll, got %d", w.Code)
	}

	// 6. Invalid option ID -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"invalid-opt"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid option, got %d", w.Code)
	}

	// 7. Malformed request (empty optionId or broken JSON) -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":""}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for empty optionId, got %d", w.Code)
	}

	// 8. Close poll via PATCH /api/polls/:id/status
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/api/polls/"+pollID+"/status", bytes.NewBufferString(`{"isActive": false}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorResp.Token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for closing poll, got %d: %s", w.Code, w.Body.String())
	}

	// 9. Closed poll rejection -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"`+opt2+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Voter-ID", "voter-bob")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for closed poll, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPollEndpoints_ConcurrentAudienceVoting(t *testing.T) {
	router, authService, _, _ := setupPollTestEnvironment()
	ctx := context.Background()

	creatorResp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "concurrent_creator@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to register creator: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/polls", bytes.NewBufferString(`{"question":"Concurrent audience vote?","options":["A","B"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorResp.Token)
	router.ServeHTTP(w, req)

	var createResp struct {
		Data models.PollResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	pollID := createResp.Data.ID
	optA := createResp.Data.Options[0].ID
	optB := createResp.Data.Options[1].ID

	concurrency := 24
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			opt := optA
			if idx%2 == 1 {
				opt = optB
			}
			rec := httptest.NewRecorder()
			vReq, _ := http.NewRequest(http.MethodPost, "/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"`+opt+`"}`))
			vReq.Header.Set("Content-Type", "application/json")
			vReq.Header.Set("X-Voter-ID", fmt.Sprintf("audience-%d", idx))
			router.ServeHTTP(rec, vReq)
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 OK for concurrent vote %d, got %d", idx, rec.Code)
			}
		}(i)
	}

	wg.Wait()

	// Verify results
	resRec := httptest.NewRecorder()
	rReq, _ := http.NewRequest(http.MethodGet, "/api/polls/"+pollID+"/results", nil)
	router.ServeHTTP(resRec, rReq)

	var results struct {
		Data models.PollResultsResponse `json:"data"`
	}
	_ = json.Unmarshal(resRec.Body.Bytes(), &results)
	if results.Data.TotalVotes != int64(concurrency) {
		t.Fatalf("expected total votes %d, got %d", concurrency, results.Data.TotalVotes)
	}
}
