package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client instance.
type RedisClient struct {
	Client *redis.Client
}

// ConnectRedis initializes a connection to Redis with short connection timeouts.
// Supports both standard host:port and full redis:// or rediss:// connection URLs.
func ConnectRedis(ctx context.Context, addr, password string, db int) (*RedisClient, error) {
	var opts *redis.Options

	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		parsedOpts, err := redis.ParseURL(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis connection URL: %w", err)
		}
		opts = parsedOpts
		if password != "" {
			opts.Password = password
		}
		if db != 0 {
			opts.DB = db
		}
	} else {
		opts = &redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}
	}

	opts.DialTimeout = 3 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Printf("[WARN] Redis ping failed at startup: %v", err)
		return &RedisClient{Client: client}, err
	}

	log.Println("[INFO] Successfully connected to Redis")
	return &RedisClient{Client: client}, nil
}

// Ping verifies that Redis is responsive.
func (r *RedisClient) Ping(ctx context.Context) error {
	if r == nil || r.Client == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.Client.Ping(pingCtx).Err()
}

// Close gracefully closes the Redis connection pool.
func (r *RedisClient) Close() error {
	if r == nil || r.Client == nil {
		return nil
	}
	log.Println("[INFO] Closing Redis client connection...")
	return r.Client.Close()
}
