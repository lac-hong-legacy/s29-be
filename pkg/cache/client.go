// pkg/cache/client.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewConfig() *Config {
	return &Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0, // default DB
	}
}

func NewClient(config *Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// Queue operations
func (c *Client) Enqueue(ctx context.Context, queueName string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return c.rdb.LPush(ctx, queueName, jsonData).Err()
}

func (c *Client) EnqueueWithDelay(ctx context.Context, queueName string, data interface{}, delay time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	score := float64(time.Now().Add(delay).Unix())
	return c.rdb.ZAdd(ctx, queueName+":delayed", redis.Z{
		Score:  score,
		Member: jsonData,
	}).Err()
}

func (c *Client) Dequeue(ctx context.Context, queueName string, timeout time.Duration) ([]byte, error) {
	result, err := c.rdb.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No items in queue
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format")
	}

	return []byte(result[1]), nil
}

// Set operations
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Hash operations
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.rdb.HGet(ctx, key, field).Result()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// List operations
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.LPush(ctx, key, values...).Err()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.RPush(ctx, key, values...).Err()
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, key).Result()
}

// Sorted set operations for delayed jobs
func (c *Client) ProcessDelayedJobs(ctx context.Context, queueName string) error {
	now := float64(time.Now().Unix())
	
	// Get jobs that are ready to be processed
	jobs, err := c.rdb.ZRangeByScore(ctx, queueName+":delayed", &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	
	if err != nil {
		return err
	}

	for _, job := range jobs {
		// Move job to main queue
		pipe := c.rdb.TxPipeline()
		pipe.LPush(ctx, queueName, job)
		pipe.ZRem(ctx, queueName+":delayed", job)
		_, err = pipe.Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Utility functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}