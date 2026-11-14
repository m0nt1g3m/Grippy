package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateToDoRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"required"`
}

type UpdateToDoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Content     string `json:"content" validate:"required"`
	IsCompleted bool   `json:"is_completed"`
}

type ToDoResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	IsCompleted bool      `json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
