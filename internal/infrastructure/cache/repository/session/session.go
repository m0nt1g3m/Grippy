package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domain_auth "Grippy/internal/domain/auth"
	"Grippy/internal/infrastructure/cache"
)

type SessionRepository struct {
	client *cache.RedisClient
}

func NewSessionRepository(client *cache.RedisClient) *SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) CreateSession(ctx context.Context, s *domain_auth.RefreshSession) error {
	key := r.buildKey(s.RefreshToken())

	model := sessionCacheModel{
		UserID:       s.UserID(),
		RefreshToken: s.RefreshToken(),
		CreatedAt:    s.CreatedAt(),
		ExpiresAt:    s.ExpiresAt(),
	}

	data, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl := time.Until(s.ExpiresAt())
	if err := r.client.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save session to redis: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetSessionByToken(ctx context.Context, token string) (*domain_auth.RefreshSession, error) {
	key := r.buildKey(token)

	resStr, err := r.client.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, domain_auth.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var model sessionCacheModel
	if err := json.Unmarshal([]byte(resStr), &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return domain_auth.NewRefreshSession(
		model.UserID,
		model.RefreshToken,
		model.CreatedAt,
		model.ExpiresAt,
	), nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, token string) error {
	key := r.buildKey(token)
	return r.client.Client.Del(ctx, key).Err()
}

func (r *SessionRepository) buildKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}
