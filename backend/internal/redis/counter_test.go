package redis_test

import (
	"context"
	"testing"

	"live-poll-backend/internal/redis"
)

func TestFormatOptionKey(t *testing.T) {
	key := redis.FormatOptionKey("poll123", "opt456")
	expected := "poll:poll123:option:opt456"
	if key != expected {
		t.Fatalf("expected key %s, got %s", expected, key)
	}

	// Handles whitespace
	trimmedKey := redis.FormatOptionKey("  poll123  ", "  opt456  ")
	if trimmedKey != expected {
		t.Fatalf("expected trimmed key %s, got %s", expected, trimmedKey)
	}
}

func TestRedisVoteCounter_UninitializedClient(t *testing.T) {
	counter := redis.NewVoteCounter(nil)
	ctx := context.Background()

	if counter.IsAvailable(ctx) {
		t.Error("expected uninitialized client to be unavailable")
	}

	_, err := counter.IncrementOption(ctx, "p1", "o1")
	if err == nil {
		t.Error("expected error when incrementing on uninitialized client")
	}

	err = counter.InitPollCounters(ctx, "p1", []string{"o1", "o2"})
	if err == nil {
		t.Error("expected error when initializing on uninitialized client")
	}

	_, err = counter.GetPollResults(ctx, "p1", []string{"o1"})
	if err == nil {
		t.Error("expected error when getting results on uninitialized client")
	}
}
