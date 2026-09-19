package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client instance.
type RedisClient struct {
	Client *redis.Client
}

// ConnectRedis initializes a connection to Redis with short connection timeouts.
func ConnectRedis(ctx context.Context, addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Printf("[WARN] Redis ping failed at startup (%s): %v", addr, err)
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
