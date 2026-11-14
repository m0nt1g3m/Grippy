package todo

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, t *ToDo) (*ToDo, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*ToDo, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
