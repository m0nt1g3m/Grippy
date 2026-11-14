package todo

import (
	"Grippy/internal/application/interfaces"
	"Grippy/internal/domain/todo"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo   todo.Repository
	logger interfaces.Logger
}

func NewService(repo todo.Repository, logger interfaces.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Save(ctx context.Context, input CreateToDoInput) (*todo.ToDo, error) {
	t, err := todo.NewTodo(input.UserID, input.Title, input.Content)

	if err != nil {
		s.logger.Error("failed to create domain todo entity", err)
		return nil, err
	}

	savedToDo, err := s.repo.Save(ctx, t)
	if err != nil {
		s.logger.Error("failed to save todo to repository", err)
		return nil, err
	}

	return savedToDo, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*todo.ToDo, error) {
	t, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) Update(ctx context.Context, input UpdateToDoInput) (*todo.ToDo, error) {
	t, err := s.repo.Get(ctx, input.ID, input.UserID)

	if err != nil {
		return nil, err
	}

	if err := t.Update(input.Title, input.Content, input.IsCompleted); err != nil {
		return nil, err
	}

	updatedToDo, err := s.repo.Save(ctx, t)

	if err != nil {
		return nil, err
	}

	return updatedToDo, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}

	return nil
}
