package user

import (
	"Grippy/internal/domain/user"

	"database/sql"
	"time"

	"github.com/google/uuid"
)

type userModel struct {
	ID           uuid.UUID      `db:"id"`
	Email        string         `db:"email"`
	Username     string         `db:"username"`
	PasswordHash string         `db:"password_hash"`
	AvatarURL    sql.NullString `db:"avatar_url"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func fromUserDomain(u *user.User) *userModel {
	return &userModel{
		ID:           u.ID(),
		Email:        u.Email(),
		Username:     u.Username(),
		PasswordHash: u.PasswordHash(),
		AvatarURL: sql.NullString{
			String: u.AvatarURL(),
			Valid:  u.AvatarURL() != "",
		},
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

func (m *userModel) toDomain() *user.User {
	var avatar string

	if m.AvatarURL.Valid {
		avatar = m.AvatarURL.String
	}

	return user.UnmarshalUserFromStorage(
		m.ID,
		m.Email,
		m.Username,
		m.PasswordHash,
		avatar,
		m.CreatedAt,
		m.UpdatedAt,
	)
}
