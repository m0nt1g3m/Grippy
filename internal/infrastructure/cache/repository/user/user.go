package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain_user "Grippy/internal/domain/user"
	"Grippy/internal/infrastructure/cache"
)

type UserRedisCache struct {
	client *cache.RedisClient
}

func NewUserRedisCache(client *cache.RedisClient) *UserRedisCache {
	return &UserRedisCache{client: client}
}

func (c *UserRedisCache) Set(ctx context.Context, u *domain_user.User, ttl time.Duration) error {
	key := c.buildKey(u.ID())

	model := userCacheModel{
		ID:           u.ID(),
		Email:        u.Email(),
		Username:     u.Username(),
		PasswordHash: u.PasswordHash(),
		AvatarURL:    u.AvatarURL(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}

	data, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal user for cache: %w", err)
	}

	if err := c.client.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save user to redis: %w", err)
	}

	return nil
}

func (c *UserRedisCache) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	key := c.buildKey(id)

	resStr, err := c.client.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var model userCacheModel
	if err := json.Unmarshal([]byte(resStr), &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached user: %w", err)
	}

	return domain_user.UnmarshalUserFromStorage(
		model.ID,
		model.Email,
		model.Username,
		model.PasswordHash,
		model.AvatarURL,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func (c *UserRedisCache) Delete(ctx context.Context, id uuid.UUID) error {
	key := c.buildKey(id)
	if err := c.client.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user from redis: %w", err)
	}
	return nil
}

func (c *UserRedisCache) buildKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}
