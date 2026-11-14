package user

import (
	"context"
	"io"
)

type AvatarStorage interface {
	Upload(ctx context.Context, folder string, filename string, file io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, fileURL string) error
}
