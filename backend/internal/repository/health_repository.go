package repository

import (
	"context"

	"live-poll-backend/internal/database"
	"live-poll-backend/internal/redis"
)

// HealthRepository defines the interface for checking storage backend availability.
type HealthRepository interface {
	PingMongoDB(ctx context.Context) error
	PingRedis(ctx context.Context) error
}

type healthRepository struct {
	mongo *database.MongoDB
	redis *redis.RedisClient
}

// NewHealthRepository creates a new instance of HealthRepository.
func NewHealthRepository(mongo *database.MongoDB, redis *redis.RedisClient) HealthRepository {
	return &healthRepository{
		mongo: mongo,
		redis: redis,
	}
}

// PingMongoDB checks the MongoDB connectivity via the database driver.
func (r *healthRepository) PingMongoDB(ctx context.Context) error {
	if r.mongo == nil {
		return database.ErrDatabaseNotInitialized
	}
	return r.mongo.Ping(ctx)
}

// PingRedis checks the Redis connectivity via the redis client.
func (r *healthRepository) PingRedis(ctx context.Context) error {
	if r.redis == nil {
		return database.ErrDatabaseNotInitialized
	}
	return r.redis.Ping(ctx)
}
