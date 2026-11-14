package user

import (
	"Grippy/internal/application/interfaces"
	"Grippy/internal/domain/user"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const userCacheTTL = 15 * time.Minute

type Service struct {
	repo    user.Repository
	cache   user.Cache
	storage user.AvatarStorage
	logger  interfaces.Logger
}

func NewService(repo user.Repository, cache user.Cache, storage user.AvatarStorage, logger interfaces.Logger) *Service {
	return &Service{
		repo:    repo,
		cache:   cache,
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	cachedUser, err := s.cache.GetByID(ctx, userID)

	if err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	u, err := s.repo.GetByID(ctx, userID)

	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, u, userCacheTTL); err != nil {
		s.logger.Warnw("failed to cache user profile", "error", err, "user_id", userID)
	}

	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*user.User, error) {
	u, err := s.repo.GetByID(ctx, input.UserID)

	if err != nil {
		return nil, err
	}

	if u.Email() != input.Email {
		if existing, err := s.repo.GetByEmail(ctx, input.Email); err == nil && existing.ID() != u.ID() {
			return nil, user.ErrUserAlreadyExists
		}
	}

	if u.Username() != input.Username {
		if existing, err := s.repo.GetByUsername(ctx, input.Username); err == nil && existing.ID() != u.ID() {
			return nil, user.ErrUsernameAlreadyExists
		}
	}

	if err := u.UpdateProfile(input.Email, input.Username); err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.Save(ctx, u)

	if err != nil {
		s.logger.Errorw("failed to save updated user profile", "error", err, "user_id", input.UserID)
		return nil, err
	}

	if err := s.cache.Delete(ctx, updatedUser.ID()); err != nil {
		s.logger.Warnw("failed to invalidate user cache", "error", err, "user_id", updatedUser.ID())
	}

	return updatedUser, nil
}

func (s *Service) UploadAvatar(ctx context.Context, input UploadAvatarInput) (*user.User, error) {
	u, err := s.repo.GetByID(ctx, input.UserID)

	if err != nil {
		return nil, err
	}

	oldAvatarURL := u.AvatarURL()

	ext := filepath.Ext(input.FileName)
	folder := fmt.Sprintf("avatars/%s", u.ID().String())
	filename := fmt.Sprintf("%s%s", uuid.NewString(), ext)

	newAvatarURL, err := s.storage.Upload(ctx, folder, filename, input.File, input.ContentType)
	if err != nil {
		s.logger.Errorw("failed to upload avatar to S3", "error", err, "user_id", u.ID())
		return nil, err
	}

	u.UpdateAvatar(newAvatarURL)
	updatedUser, err := s.repo.Save(ctx, u)

	if err != nil {
		_ = s.storage.Delete(ctx, newAvatarURL)
		s.logger.Errorw("failed to save avatar url to db", "error", err, "user_id", u.ID())
		return nil, err
	}

	if oldAvatarURL != "" && oldAvatarURL != newAvatarURL {
		if err := s.storage.Delete(ctx, oldAvatarURL); err != nil {
			s.logger.Warnw("failed to delete old avatar from S3", "error", err, "old_url", oldAvatarURL)
		}
	}

	_ = s.cache.Delete(ctx, updatedUser.ID())

	return updatedUser, nil
}
