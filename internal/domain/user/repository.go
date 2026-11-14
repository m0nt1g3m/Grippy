package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, u *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
