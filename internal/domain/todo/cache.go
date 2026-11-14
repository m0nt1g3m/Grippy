package todo

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Cache interface {
	Set(ctx context.Context, t *ToDo, ttl time.Duration) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*ToDo, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
