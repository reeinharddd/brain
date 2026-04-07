package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines cache operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Flush(ctx context.Context) error
	Health(ctx context.Context) error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss: %s", key)
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes keys from cache
func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return rc.client.Del(ctx, keys...).Err()
}

// Flush clears all cache
func (rc *RedisCache) Flush(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// Health checks Redis connectivity
func (rc *RedisCache) Health(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// NoopCache is a no-op cache implementation (used when Redis is unavailable)
type NoopCache struct{}

// Get always returns cache miss
func (nc *NoopCache) Get(ctx context.Context, key string, dest interface{}) error {
	return fmt.Errorf("cache miss: %s (no-op cache)", key)
}

// Set does nothing
func (nc *NoopCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (nc *NoopCache) Delete(ctx context.Context, keys ...string) error {
	return nil
}

// Flush does nothing
func (nc *NoopCache) Flush(ctx context.Context) error {
	return nil
}

// Health always returns unavailable
func (nc *NoopCache) Health(ctx context.Context) error {
	return fmt.Errorf("no-op cache (Redis unavailable)")
}

// NewCache creates a cache instance (Redis or no-op)
func NewCache(redisAddr string) Cache {
	if redisAddr == "" {
		return &NoopCache{}
	}

	cache, err := NewRedisCache(redisAddr)
	if err != nil {
		return &NoopCache{}
	}

	return cache
}
