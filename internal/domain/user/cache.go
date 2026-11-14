package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Cache interface {
	Set(ctx context.Context, u *User, ttl time.Duration) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
