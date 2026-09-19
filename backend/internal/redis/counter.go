package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrRedisNotConnected is returned when Redis operations are attempted on an uninitialized client.
	ErrRedisNotConnected = errors.New("redis client is not connected")
)

// VoteCounter defines operations for maintaining live poll counts in Redis.
type VoteCounter interface {
	InitPollCounters(ctx context.Context, pollID string, optionIDs []string) error
	IncrementOption(ctx context.Context, pollID, optionID string) (int64, error)
	GetOptionCount(ctx context.Context, pollID, optionID string) (int64, error)
	GetPollResults(ctx context.Context, pollID string, optionIDs []string) (map[string]int64, error)
	SetOptionCount(ctx context.Context, pollID, optionID string, count int64) error
	DeletePollCounters(ctx context.Context, pollID string, optionIDs []string) error
	IsAvailable(ctx context.Context) bool
}

type redisVoteCounter struct {
	client *RedisClient
}

// NewVoteCounter instantiates a new VoteCounter backed by Redis.
func NewVoteCounter(client *RedisClient) VoteCounter {
	return &redisVoteCounter{client: client}
}

// cleanKeyPart strips any control characters, spaces, colons, or wildcards to prevent key manipulation.
func cleanKeyPart(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FormatOptionKey constructs a predictable, sanitized Redis key for a poll option.
// Key format: poll:{pollID}:option:{optionID}
func FormatOptionKey(pollID, optionID string) string {
	return fmt.Sprintf("poll:%s:option:%s", cleanKeyPart(pollID), cleanKeyPart(optionID))
}

// IsAvailable checks if the Redis client is reachable.
func (c *redisVoteCounter) IsAvailable(ctx context.Context) bool {
	if c.client == nil || c.client.Client == nil {
		return false
	}
	return c.client.Ping(ctx) == nil
}

// InitPollCounters sets option counters to 0 using SETNX so existing counts are preserved.
func (c *redisVoteCounter) InitPollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	if c.client == nil || c.client.Client == nil {
		return ErrRedisNotConnected
	}

	pipe := c.client.Client.Pipeline()
	for _, optID := range optionIDs {
		key := FormatOptionKey(pollID, optID)
		// SetNX sets value only if key does not exist yet
		pipe.SetNX(ctx, key, 0, 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize poll counters in Redis: %w", err)
	}

	return nil
}

// IncrementOption executes an atomic Redis INCR command on the option key.
func (c *redisVoteCounter) IncrementOption(ctx context.Context, pollID, optionID string) (int64, error) {
	if c.client == nil || c.client.Client == nil {
		return 0, ErrRedisNotConnected
	}

	key := FormatOptionKey(pollID, optionID)
	newCount, err := c.client.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis INCR failed on key %s: %w", key, err)
	}

	return newCount, nil
}

// GetOptionCount reads the current count for an individual option.
func (c *redisVoteCounter) GetOptionCount(ctx context.Context, pollID, optionID string) (int64, error) {
	if c.client == nil || c.client.Client == nil {
		return 0, ErrRedisNotConnected
	}

	key := FormatOptionKey(pollID, optionID)
	val, err := c.client.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get option count for %s: %w", key, err)
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid counter value stored in %s: %w", key, err)
	}

	return count, nil
}

// GetPollResults fetches live counts for all given options using a fast MGet pipeline.
func (c *redisVoteCounter) GetPollResults(ctx context.Context, pollID string, optionIDs []string) (map[string]int64, error) {
	results := make(map[string]int64, len(optionIDs))
	if len(optionIDs) == 0 {
		return results, nil
	}

	if c.client == nil || c.client.Client == nil {
		return nil, ErrRedisNotConnected
	}

	keys := make([]string, len(optionIDs))
	for i, optID := range optionIDs {
		keys[i] = FormatOptionKey(pollID, optID)
	}

	values, err := c.client.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis MGet failed for poll %s: %w", pollID, err)
	}

	for i, optID := range optionIDs {
		val := values[i]
		if val == nil {
			// Key is missing or expired in Redis
			results[optID] = -1 // Marker for missing key to trigger reconciliation
			continue
		}

		strVal, ok := val.(string)
		if !ok {
			results[optID] = 0
			continue
		}

		count, parseErr := strconv.ParseInt(strVal, 10, 64)
		if parseErr != nil {
			results[optID] = 0
			continue
		}
		results[optID] = count
	}

	return results, nil
}

// SetOptionCount manually sets the counter value for an option (e.g. for reconciliation).
func (c *redisVoteCounter) SetOptionCount(ctx context.Context, pollID, optionID string, count int64) error {
	if c.client == nil || c.client.Client == nil {
		return ErrRedisNotConnected
	}

	key := FormatOptionKey(pollID, optionID)
	return c.client.Client.Set(ctx, key, count, 0).Err()
}

// DeletePollCounters removes all Redis keys associated with a deleted poll.
func (c *redisVoteCounter) DeletePollCounters(ctx context.Context, pollID string, optionIDs []string) error {
	if c.client == nil || c.client.Client == nil {
		return nil
	}

	if len(optionIDs) == 0 {
		return nil
	}

	keys := make([]string, len(optionIDs))
	for i, optID := range optionIDs {
		keys[i] = FormatOptionKey(pollID, optID)
	}

	return c.client.Client.Del(ctx, keys...).Err()
}
