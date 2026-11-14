package todo

import "github.com/google/uuid"

// Parameters for creating todo
type CreateToDoInput struct {
	UserID  uuid.UUID
	Title   string
	Content string
}

// Parameters for updaing todo
type UpdateToDoInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Content     string
	IsCompleted bool
}
