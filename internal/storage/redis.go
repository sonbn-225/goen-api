package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(redisURL string) *Redis {
	if redisURL == "" {
		return nil
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil
	}
	return &Redis{client: redis.NewClient(opts)}
}

func (r *Redis) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Probe(ctx context.Context) (map[string]any, error) {
	if r == nil || r.client == nil {
		return nil, nil
	}

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	key := "goen_api_probe"
	value := time.Now().UTC().Format(time.RFC3339Nano)
	if err := r.client.Set(ctx, key, value, 30*time.Second).Err(); err != nil {
		return nil, fmt.Errorf("redis set: %w", err)
	}
	got, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	return map[string]any{
		"ping":  "PONG",
		"key":   key,
		"value": got,
	}, nil
}

func (r *Redis) XAdd(ctx context.Context, stream string, values map[string]any) (string, error) {
	if r == nil || r.client == nil {
		return "", errors.New("redis not ready")
	}
	if stream == "" {
		return "", errors.New("stream is required")
	}
	return r.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: values}).Result()
}

func (r *Redis) Close() {
	if r == nil || r.client == nil {
		return
	}
	_ = r.client.Close()
	r.client = nil
}
