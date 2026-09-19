package services_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/repository"
	"live-poll-backend/internal/services"
)

type mockPollRepo struct {
	mu        sync.Mutex
	pollsByID map[string]*models.Poll
	votes     map[string]bool
}

func newMockPollRepo() *mockPollRepo {
	return &mockPollRepo{
		pollsByID: make(map[string]*models.Poll),
		votes:     make(map[string]bool),
	}
}

func (m *mockPollRepo) Create(ctx context.Context, poll *models.Poll) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if poll.ID.IsZero() {
		poll.ID = bson.NewObjectID()
	}
	poll.CreatedAt = time.Now().UTC()
	poll.UpdatedAt = time.Now().UTC()
	poll.IsActive = true
	m.pollsByID[poll.ID.Hex()] = poll
	return nil
}

func (m *mockPollRepo) FindByID(ctx context.Context, id string) (*models.Poll, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pollsByID[id]
	if !exists {
		return nil, repository.ErrPollNotFound
	}
	return p, nil
}

func (m *mockPollRepo) FindByCreatorID(ctx context.Context, creatorID string) ([]*models.Poll, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	polls := make([]*models.Poll, 0)
	for _, p := range m.pollsByID {
		if p.CreatorID == creatorID {
			polls = append(polls, p)
		}
	}
	return polls, nil
}

func (m *mockPollRepo) Delete(ctx context.Context, id string, creatorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pollsByID[id]
	if !exists {
		return repository.ErrPollNotFound
	}
	if p.CreatorID != creatorID {
		return repository.ErrUnauthorizedAction
	}
	delete(m.pollsByID, id)
	return nil
}

func (m *mockPollRepo) SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pollsByID[id]
	if !exists {
		return repository.ErrPollNotFound
	}
	if p.CreatorID != creatorID {
		return repository.ErrUnauthorizedAction
	}
	p.IsActive = isActive
	return nil
}

func (m *mockPollRepo) EnsureIndexes(ctx context.Context) error {
	return nil
}

func (m *mockPollRepo) RecordVote(ctx context.Context, vote *models.Vote) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pollsByID[vote.PollID]
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

func (m *mockPollRepo) HasVoterVoted(ctx context.Context, pollID, voterID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if voterID == "" {
		return false, nil
	}
	return m.votes[pollID+":"+voterID], nil
}

type mockVoteCounter struct {
	mu     sync.Mutex
	counts map[string]int64
}

func newMockVoteCounter() *mockVoteCounter {
	return &mockVoteCounter{counts: make(map[string]int64)}
}

func (c *mockVoteCounter) InitPollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, optID := range optionIDs {
		key := pollID + ":" + optID
		if _, exists := c.counts[key]; !exists {
			c.counts[key] = 0
		}
	}
	return nil
}

func (c *mockVoteCounter) IncrementOption(ctx context.Context, pollID, optionID string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := pollID + ":" + optionID
	c.counts[key]++
	return c.counts[key], nil
}

func (c *mockVoteCounter) GetOptionCount(ctx context.Context, pollID, optionID string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[pollID+":"+optionID], nil
}

func (c *mockVoteCounter) GetPollResults(ctx context.Context, pollID string, optionIDs []string) (map[string]int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := make(map[string]int64)
	for _, optID := range optionIDs {
		if val, exists := c.counts[pollID+":"+optID]; exists {
			res[optID] = val
		} else {
			res[optID] = -1
		}
	}
	return res, nil
}

func (c *mockVoteCounter) SetOptionCount(ctx context.Context, pollID, optionID string, count int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[pollID+":"+optionID] = count
	return nil
}

func (c *mockVoteCounter) DeletePollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, optID := range optionIDs {
		delete(c.counts, pollID+":"+optID)
	}
	return nil
}

func (c *mockVoteCounter) IsAvailable(ctx context.Context) bool {
	return true
}

func TestPollService_CreatePoll_Valid(t *testing.T) {
	repo := newMockPollRepo()
	counter := newMockVoteCounter()
	service := services.NewPollService(repo, counter, nil)
	ctx := context.Background()

	req := models.CreatePollRequest{
		Question: "What is your favorite programming language?",
		Options:  []string{"Go", "Rust", "TypeScript"},
	}

	resp, err := service.CreatePoll(ctx, "creator-123", req)
	if err != nil {
		t.Fatalf("unexpected error creating poll: %v", err)
	}

	if resp.ID == "" || !resp.IsActive {
		t.Errorf("expected generated ID and active poll, got ID: %s, isActive: %v", resp.ID, resp.IsActive)
	}
}

func TestPollService_Vote_Comprehensive(t *testing.T) {
	repo := newMockPollRepo()
	counter := newMockVoteCounter()
	service := services.NewPollService(repo, counter, nil)
	ctx := context.Background()

	created, err := service.CreatePoll(ctx, "creator-1", models.CreatePollRequest{
		Question: "Which cloud provider?",
		Options:  []string{"AWS", "GCP", "Azure"},
	})
	if err != nil {
		t.Fatalf("failed to create poll: %v", err)
	}

	opt1 := created.Options[0].ID
	opt2 := created.Options[1].ID

	// 1. Valid vote
	res, err := service.Vote(ctx, created.ID, opt1, "voter-1", "")
	if err != nil {
		t.Fatalf("expected successful vote, got: %v", err)
	}
	if res.TotalVotes != 1 || res.Options[0].VoteCount != 1 {
		t.Errorf("expected 1 vote for opt1, got: %+v", res)
	}

	// 2. Duplicate vote from same voter -> ErrAlreadyVoted
	_, err = service.Vote(ctx, created.ID, opt2, "voter-1", "")
	if !errors.Is(err, services.ErrAlreadyVoted) {
		t.Errorf("expected ErrAlreadyVoted for duplicate voter, got: %v", err)
	}

	// 3. Invalid poll ID
	_, err = service.Vote(ctx, "nonexistent-poll", opt1, "voter-2", "")
	if !errors.Is(err, services.ErrPollNotFound) {
		t.Errorf("expected ErrPollNotFound for nonexistent poll, got: %v", err)
	}

	// 4. Invalid option ID
	_, err = service.Vote(ctx, created.ID, "fake-option", "voter-3", "")
	if !errors.Is(err, services.ErrInvalidOption) {
		t.Errorf("expected ErrInvalidOption for fake option, got: %v", err)
	}

	// 5. Closed poll cannot receive votes
	err = service.SetPollStatus(ctx, created.ID, "creator-1", false)
	if err != nil {
		t.Fatalf("failed to close poll: %v", err)
	}
	_, err = service.Vote(ctx, created.ID, opt2, "voter-4", "")
	if !errors.Is(err, services.ErrPollClosed) {
		t.Errorf("expected ErrPollClosed for closed poll, got: %v", err)
	}
}

func TestPollService_ConcurrentVoting(t *testing.T) {
	repo := newMockPollRepo()
	counter := newMockVoteCounter()
	service := services.NewPollService(repo, counter, nil)
	ctx := context.Background()

	created, err := service.CreatePoll(ctx, "creator-concurrent", models.CreatePollRequest{
		Question: "Concurrent test question?",
		Options:  []string{"Option Alpha", "Option Beta"},
	})
	if err != nil {
		t.Fatalf("failed to create poll: %v", err)
	}

	optA := created.Options[0].ID
	optB := created.Options[1].ID

	concurrency := 30
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			opt := optA
			if idx%2 == 1 {
				opt = optB
			}
			voterID := fmt.Sprintf("voter-conc-%d", idx)
			_, voteErr := service.Vote(ctx, created.ID, opt, voterID, "")
			if voteErr != nil {
				t.Errorf("concurrent vote failed for index %d: %v", idx, voteErr)
			}
		}(i)
	}

	wg.Wait()

	results, err := service.GetPollResults(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get results after concurrent voting: %v", err)
	}

	if results.TotalVotes != int64(concurrency) {
		t.Fatalf("expected %d total votes, got %d", concurrency, results.TotalVotes)
	}

	half := int64(concurrency / 2)
	if results.Options[0].VoteCount != half || results.Options[1].VoteCount != half {
		t.Errorf("expected %d votes each, got opt0=%d, opt1=%d", half, results.Options[0].VoteCount, results.Options[1].VoteCount)
	}
}
