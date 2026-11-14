package todo

import (
	"Grippy/internal/domain/todo"
	"context"
	"errors"

	"github.com/google/uuid"

	"Grippy/internal/infrastructure/database"
)

type ToDoRepository struct {
	db *database.Postgres
}

func NewToDoRepository(db *database.Postgres) *ToDoRepository {
	return &ToDoRepository{
		db: db,
	}
}

func (r *ToDoRepository) Save(ctx context.Context, t *todo.ToDo) (*todo.ToDo, error) {
	dbModel := fromDomain(t)

	query := `
			INSERT INTO todo_schema.todos (id, user_id, title, content, is_completed, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, content = EXCLUDED.content, is_completed = EXCLUDED.is_completed, updated_at = EXCLUDED.updated_at RETURNING id, user_id, title, content, is_completed, created_at, updated_at`
	var resultModel todoModel

	err := r.db.Pool.QueryRow(ctx, query,
		dbModel.ID,
		dbModel.UserID,
		dbModel.Title,
		dbModel.Content,
		dbModel.IsCompleted,
		dbModel.CreatedAt,
		dbModel.UpdatedAt,
	).Scan(
		&resultModel.ID,
		&resultModel.UserID,
		&resultModel.Title,
		&resultModel.Content,
		&resultModel.IsCompleted,
		&resultModel.CreatedAt,
		&resultModel.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return resultModel.toDomain()
}

func (r *ToDoRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*todo.ToDo, error) {
	query := `
			SELECT id, user_id, title, content, is_completed, created_at, updated_at
			FROM todo_schema.todos
			WHERE id = $1 AND user_id = $2`
	var dbModel todoModel

	err := r.db.Pool.QueryRow(ctx, query, id, userID).Scan(
		&dbModel.ID,
		&dbModel.UserID,
		&dbModel.Title,
		&dbModel.Content,
		&dbModel.IsCompleted,
		&dbModel.CreatedAt,
		&dbModel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, database.ErrNoRows) {
			return nil, todo.ErrToDoNotFound
		}
		return nil, err
	}

	return dbModel.toDomain()
}

func (r *ToDoRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
			DELETE FROM todo_schema.todos
			WHERE id = $1 AND user_id = $2`
	result, err := r.db.Pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return todo.ErrToDoNotFound
	}

	return nil
}
