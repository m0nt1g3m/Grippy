package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	srv "Grippy/internal/application/user"
	domain "Grippy/internal/domain/user"
	"Grippy/internal/infrastructure/logging"
	"Grippy/internal/transport/http/dto"
	"Grippy/internal/transport/http/router"
)

const maxAvatarSize = 5 << 20 // 5 MB

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, input srv.UpdateProfileInput) (*domain.User, error)
	UploadAvatar(ctx context.Context, input srv.UploadAvatarInput) (*domain.User, error)
}

type UserHandler struct {
	service UserService
	logger  *logging.Logger
}

func NewUserHandler(service UserService, logger *logging.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) RegisterRoutes(r *router.Router) {
	r.Get("/profile", h.GetProfile)
	r.Put("/profile", h.UpdateProfile)
	r.Post("/avatar", h.UploadAvatar)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, h.mapToResponse(u))
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := srv.UpdateProfileInput{
		UserID:   userID,
		Email:    req.Email,
		Username: req.Username,
	}

	u, err := h.service.UpdateProfile(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, h.mapToResponse(u))
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		respondWithError(w, http.StatusBadRequest, "file is too large (max 5MB)")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to read file")
		return
	}
	contentType := http.DetectContentType(buf[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		respondWithError(w, http.StatusBadRequest, "only jpeg, png and webp images are allowed")
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	input := srv.UploadAvatarInput{
		UserID:      userID,
		File:        file,
		Size:        header.Size,
		FileName:    header.Filename,
		ContentType: contentType,
	}

	u, err := h.service.UploadAvatar(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, h.mapToResponse(u))
}

func (h *UserHandler) mapToResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID(),
		Email:     u.Email(),
		Username:  u.Username(),
		AvatarURL: u.AvatarURL(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

func (h *UserHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		respondWithError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrUserAlreadyExists), errors.Is(err, domain.ErrUsernameAlreadyExists):
		respondWithError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrEmptyEmail),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrEmptyUsername),
		errors.Is(err, domain.ErrUsernameTooLong):
		respondWithError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Errorf("unexpected error inside user handler: %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")
	}
}
