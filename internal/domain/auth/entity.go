package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrSessionNotFound = errors.New("refresh session not found")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	ID        uuid.UUID
	Username  string
	ExpiresAt time.Time
}

type RefreshSession struct {
	userID       uuid.UUID
	refreshToken string
	createdAt    time.Time
	expiresdAt   time.Time
}

func NewRefreshSession(userID uuid.UUID, refreshToken string, createdAt, expiresdAt time.Time) *RefreshSession {
	return &RefreshSession{
		userID:       userID,
		refreshToken: refreshToken,
		createdAt:    createdAt,
		expiresdAt:   expiresdAt,
	}
}

func (s *RefreshSession) IsExpired() bool {
	return time.Now().After(s.expiresdAt)
}

// Getters
func (s *RefreshSession) UserID() uuid.UUID    { return s.userID }
func (s *RefreshSession) RefreshToken() string { return s.refreshToken }
func (s *RefreshSession) CreatedAt() time.Time { return s.createdAt }
func (s *RefreshSession) ExpiresAt() time.Time { return s.expiresdAt }
