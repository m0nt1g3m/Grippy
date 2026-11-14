package user

import (
	"io"

	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	UserID   uuid.UUID
	Email    string
	Username string
}

type UploadAvatarInput struct {
	UserID      uuid.UUID
	File        io.Reader
	Size        int64
	FileName    string
	ContentType string
}
