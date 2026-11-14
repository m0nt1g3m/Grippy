package dto

import (
	"time"

	"github.com/google/uuid"
)

type UpdateProfileRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
