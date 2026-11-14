package todo

import (
	"Grippy/internal/domain/todo"
	"time"

	"github.com/google/uuid"
)

type todoModel struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	Title       string    `db:"title"`
	Content     string    `db:"content"`
	IsCompleted bool      `db:"is_completed"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func fromDomain(t *todo.ToDo) *todoModel {
	return &todoModel{
		ID:          t.ID(),
		UserID:      t.UserID(),
		Title:       t.Title(),
		Content:     t.Content(),
		IsCompleted: t.IsCompleted(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

func (m *todoModel) toDomain() (*todo.ToDo, error) {
	return todo.UnmarshalTodoFromDB(
		m.ID,
		m.UserID,
		m.Title,
		m.Content,
		m.IsCompleted,
		m.CreatedAt,
		m.UpdatedAt,
	), nil
}
