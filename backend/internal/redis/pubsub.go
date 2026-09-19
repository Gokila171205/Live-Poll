package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// EventPublisher defines publishing capabilities for real-time poll updates.
type EventPublisher interface {
	Publish(ctx context.Context, pollID string, payload []byte) error
}

// EventSubscriber defines subscription capabilities for listening to Redis Pub/Sub events.
type EventSubscriber interface {
	Subscribe(ctx context.Context, handler func(pollID string, payload []byte)) error
	Close() error
}

// PubSubService combines publishing and subscription capabilities.
type PubSubService interface {
	EventPublisher
	EventSubscriber
}

type redisPubSubService struct {
	client     *RedisClient
	pubsub     *goredis.PubSub
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	localSubs  []func(pollID string, payload []byte)
}

// NewPubSubService creates a new PubSubService backed by Redis.
func NewPubSubService(client *RedisClient) PubSubService {
	return &redisPubSubService{
		client:    client,
		localSubs: make([]func(pollID string, payload []byte), 0),
	}
}

// cleanChannelPart strips control characters, spaces, colons, or wildcards to prevent channel injection.
func cleanChannelPart(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FormatChannelName generates a structured, sanitized Redis Pub/Sub channel for a poll.
func FormatChannelName(pollID string) string {
	return fmt.Sprintf("poll:events:%s", cleanChannelPart(pollID))
}

// Publish sends an update message to the Redis Pub/Sub channel for a specific poll.
// When Redis is connected, it publishes to the Redis channel "poll:events:<pollID>".
// If Redis is not connected (e.g. mock/test mode), it dispatches to local subscribers.
func (s *redisPubSubService) Publish(ctx context.Context, pollID string, payload []byte) error {
	if s.client != nil && s.client.Client != nil {
		channel := FormatChannelName(pollID)
		err := s.client.Client.Publish(ctx, channel, payload).Err()
		if err != nil {
			return fmt.Errorf("failed to publish to Redis channel %s: %w", channel, err)
		}
		return nil
	}

	// Fallback for tests/environments without an external Redis instance
	s.mu.RLock()
	subs := make([]func(pollID string, payload []byte), len(s.localSubs))
	copy(subs, s.localSubs)
	s.mu.RUnlock()

	for _, sub := range subs {
		sub(pollID, payload)
	}

	return nil
}

// Subscribe listens to all poll event channels (pattern "poll:events:*") with resilient auto-reconnect.
func (s *redisPubSubService) Subscribe(ctx context.Context, handler func(pollID string, payload []byte)) error {
	s.mu.Lock()
	s.localSubs = append(s.localSubs, handler)

	if s.client == nil || s.client.Client == nil {
		s.mu.Unlock()
		return nil
	}

	subCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel
	s.pubsub = s.client.Client.PSubscribe(subCtx, "poll:events:*")
	s.mu.Unlock()

	// Listen for messages in background goroutine with automatic reconnection on disconnect
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] Redis Pub/Sub subscriber panic recovered: %v", r)
			}
		}()

		for {
			select {
			case <-subCtx.Done():
				return
			default:
			}

			s.mu.RLock()
			ps := s.pubsub
			s.mu.RUnlock()

			if ps == nil {
				return
			}

			ch := ps.Channel()
			closed := false
			for !closed {
				select {
				case <-subCtx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						closed = true
						break
					}
					// Extract pollID from channel "poll:events:<pollID>"
					parts := strings.Split(msg.Channel, ":")
					var pollID string
					if len(parts) >= 3 {
						pollID = parts[2]
					}
					handler(pollID, []byte(msg.Payload))
				}
			}

			// Backoff and re-subscribe if context not cancelled
			select {
			case <-subCtx.Done():
				return
			case <-time.After(2 * time.Second):
				s.mu.Lock()
				if s.client != nil && s.client.Client != nil {
					if s.pubsub != nil {
						_ = s.pubsub.Close()
					}
					s.pubsub = s.client.Client.PSubscribe(subCtx, "poll:events:*")
					log.Println("[INFO] Redis Pub/Sub pattern subscriber re-established connection on poll:events:*")
				}
				s.mu.Unlock()
			}
		}
	}()

	log.Println("[INFO] Redis Pub/Sub pattern subscriber listening on poll:events:*")
	return nil
}

// Close gracefully closes the Redis pubsub listener connection and cancels background goroutines.
func (s *redisPubSubService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.localSubs = nil
	if s.pubsub != nil {
		return s.pubsub.Close()
	}
	return nil
}
