package cache

import (
	"Grippy/internal/infrastructure/logging"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
	logger *logging.Logger
}

const Nil = redis.Nil

func NewRedisClient(host string, port string, logger *logging.Logger) (*RedisClient, error) {
	address := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:         address,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisClient{
		Client: client,
		logger: logger,
	}, nil
}

func (client *RedisClient) CloseClient() {
	err := client.Close()
	if err != nil {
		client.logger.Errorw("close redis client", "err", err)
		return
	}
	client.logger.Info("Redis client closed")
}
