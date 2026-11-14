package user

import (
	"Grippy/internal/domain/user"
	"Grippy/internal/infrastructure/database"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *database.Postgres
}

func NewUserRepository(db *database.Postgres) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) (*user.User, error) {
	dbModel := fromUserDomain(u)

	query := `
		INSERT INTO user_schema.users (id, email, username, password_hash, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email,
			username = EXCLUDED.username, 
		    password_hash = COALESCE(NULLIF(EXCLUDED.password_hash, ''), users.password_hash),
		    avatar_url = EXCLUDED.avatar_url,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, email, username, password_hash, avatar_url, created_at, updated_at;
	`

	var res userModel

	err := r.db.Pool.QueryRow(ctx, query,
		dbModel.ID,
		dbModel.Email,
		dbModel.Username,
		dbModel.PasswordHash,
		dbModel.AvatarURL,
		dbModel.CreatedAt,
		dbModel.UpdatedAt,
	).Scan(
		&res.ID,
		&res.Email,
		&res.Username,
		&res.PasswordHash,
		&res.AvatarURL,
		&res.CreatedAt,
		&res.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, user.ErrUserAlreadyExists
			}

			if strings.Contains(pgErr.ConstraintName, "username") {
				return nil, user.ErrUsernameAlreadyExists
			}
		}

		return nil, err
	}

	return res.toDomain(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM user_schema.users
		WHERE id = $1
	`

	return r.getOne(ctx, query, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM user_schema.users
		WHERE email = $1
	`

	return r.getOne(ctx, query, email)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `SELECT id, email, username, password_hash, avatar_url, created_at, updated_at
		FROM user_schema.users
		WHERE username = $1
	`

	return r.getOne(ctx, query, username)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_schema.users WHERE id = $1`

	res, err := r.db.Pool.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) getOne(ctx context.Context, query string, arg any) (*user.User, error) {
	var m userModel

	err := r.db.Pool.QueryRow(ctx, query, arg).Scan(
		&m.ID,
		&m.Email,
		&m.Username,
		&m.PasswordHash,
		&m.AvatarURL,
		&m.CreatedAt,
		&m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return m.toDomain(), nil
}
