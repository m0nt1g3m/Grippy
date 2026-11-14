package todo

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ToDo struct {
	id          uuid.UUID
	userID      uuid.UUID
	title       string
	content     string
	isCompleted bool
	createdAt   time.Time
	updatedAt   time.Time
}

var (
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrEmptyContent = errors.New("content cannot be empty")
	ErrToDoNotFound = errors.New("todo item not found")
)

func NewTodo(userID uuid.UUID, title, content string) (*ToDo, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	if content == "" {
		return nil, ErrEmptyContent
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	return &ToDo{
		id:        id,
		userID:    userID,
		title:     title,
		content:   content,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

func (t *ToDo) Update(title, content string, isCompleted bool) error {
	if title == "" {
		return ErrEmptyTitle
	}

	if content == "" {
		return ErrEmptyContent
	}

	t.title = title
	t.content = content
	t.isCompleted = isCompleted
	t.updatedAt = time.Now()
	return nil
}

func UnmarshalTodoFromDB(id uuid.UUID, userID uuid.UUID, title, content string, isCompleted bool, createdAt time.Time, updatedAt time.Time) *ToDo {
	return &ToDo{
		id:          id,
		userID:      userID,
		title:       title,
		content:     content,
		isCompleted: isCompleted,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Getters
func (t *ToDo) ID() uuid.UUID        { return t.id }
func (t *ToDo) UserID() uuid.UUID    { return t.userID }
func (t *ToDo) Title() string        { return t.title }
func (t *ToDo) Content() string      { return t.content }
func (t *ToDo) IsCompleted() bool    { return t.isCompleted }
func (t *ToDo) CreatedAt() time.Time { return t.createdAt }
func (t *ToDo) UpdatedAt() time.Time { return t.updatedAt }
