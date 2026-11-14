package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id           uuid.UUID
	avatarURL    string
	email        string
	username     string
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
}

var (
	ErrEmptyEmail            = errors.New("email cannot be empty")
	ErrEmptyUsername         = errors.New("username cannot be empty")
	ErrEmptyPassword         = errors.New("password cannot be empty")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrUsernameTooLong       = errors.New("username is too long")
	ErrWeakPassword          = errors.New("password must be at least 6 characters long")
)

func NewUser(email, username, passwordHash string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	if email == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidEmail
	}

	if username == "" || len(username) > 50 {
		return nil, ErrEmptyUsername
	}

	if len(passwordHash) < 6 {
		return nil, ErrWeakPassword
	}

	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		id:           id,
		email:        email,
		username:     username,
		passwordHash: passwordHash,
		avatarURL:    "",
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (u *User) UpdateProfile(email, username string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)

	if email == "" {
		return ErrEmptyEmail
	}

	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	if username == "" {
		return ErrEmptyUsername
	}

	if len(username) > 50 {
		return ErrUsernameTooLong
	}

	u.email = email
	u.username = username
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateAvatar(avatarURL string) {
	u.avatarURL = strings.TrimSpace(avatarURL)
	u.updatedAt = time.Now()
}

func UnmarshalUserFromStorage(
	id uuid.UUID,
	email string,
	username string,
	passwordHash string,
	avatarURL string,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		email:        email,
		username:     username,
		passwordHash: passwordHash,
		avatarURL:    avatarURL,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Getters
func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) AvatarURL() string    { return u.avatarURL }
func (u *User) Email() string        { return u.email }
func (u *User) Username() string     { return u.username }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
