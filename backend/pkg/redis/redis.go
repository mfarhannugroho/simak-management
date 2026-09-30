package redisclient

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"simak-backend/config"
)

// Client is the global Redis client instance.
var Client *redis.Client

// Connect initialises and returns the Redis client.
func Connect(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Printf("⚠️ Failed to connect to Redis: %v", err)
		return nil
	}

	Client = client
	log.Println("Redis connected successfully")
	return client
}

// Ping checks whether the Redis connection is alive.
func Ping(client *redis.Client) error {
	if client == nil {
		return redis.ErrClosed
	}
	return client.Ping(context.Background()).Err()
}
