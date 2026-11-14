package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	srv "Grippy/internal/application/auth"
	domain_auth "Grippy/internal/domain/auth"
	domain_user "Grippy/internal/domain/user"
	"Grippy/internal/infrastructure/logging"
	"Grippy/internal/transport/http/dto"
	"Grippy/internal/transport/http/router"
)

type AuthService interface {
	SignUp(ctx context.Context, input srv.SignUpInput) (*domain_user.User, error)
	SignIn(ctx context.Context, input srv.SignInInput) (*domain_auth.TokenPair, error)
	Refresh(ctx context.Context, input srv.RefreshInput) (*domain_auth.TokenPair, error)
	Logout(ctx context.Context, input srv.LogoutInput) error
	ValidateRefreshToken(ctx context.Context, tokenStr string) error
}

type AuthHandler struct {
	service    AuthService
	logger     *logging.Logger
	refreshTTL time.Duration
	isSecure   bool
}

func NewAuthHandler(service AuthService, logger *logging.Logger, refreshTTL time.Duration, isSecure bool) *AuthHandler {
	return &AuthHandler{
		service:    service,
		logger:     logger,
		refreshTTL: refreshTTL,
		isSecure:   isSecure,
	}
}

func (h *AuthHandler) RegisterRoutes(r *router.Router) {
	r.Post("/sign_up", h.SignUp)
	r.Post("/sign_in", h.SignIn)
	r.Post("/refresh_token", h.Refresh)
	r.Post("/logout", h.Logout)
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if h.checkAlreadySignedIn(w, r) {
		return
	}

	var req dto.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := srv.SignUpInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	u, err := h.service.SignUp(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	resp := dto.UserResponse{
		ID:        u.ID(),
		Email:     u.Email(),
		Username:  u.Username(),
		AvatarURL: u.AvatarURL(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if h.checkAlreadySignedIn(w, r) {
		return
	}

	var req dto.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := srv.SignInInput{
		Email:    req.Email,
		Password: req.Password,
	}

	tokens, err := h.service.SignIn(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken)
	respondWithJSON(w, http.StatusOK, map[string]string{"access_token": tokens.AccessToken})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token cookie")
		return
	}

	tokens, err := h.service.Refresh(r.Context(), srv.RefreshInput{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.clearRefreshCookie(w)
		h.handleDomainError(w, err)
		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken)
	respondWithJSON(w, http.StatusOK, map[string]string{"access_token": tokens.AccessToken})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		respondWithJSON(w, http.StatusOK, map[string]string{"message": "already logged out"})
		return
	}

	if err := h.service.Logout(r.Context(), srv.LogoutInput{RefreshToken: cookie.Value}); err != nil {
		h.logger.Errorf("logout service error: %v", err)
	}

	h.clearRefreshCookie(w)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) checkAlreadySignedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		if err := h.service.ValidateRefreshToken(r.Context(), cookie.Value); err == nil {
			respondWithError(w, http.StatusConflict, "already signed in, please logout first")
			return true
		}
		h.clearRefreshCookie(w)
	}
	return false
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(h.refreshTTL),
		HttpOnly: true,
		Secure:   h.isSecure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth",
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain_user.ErrInvalidCredentials),
		errors.Is(err, domain_auth.ErrInvalidToken),
		errors.Is(err, domain_auth.ErrSessionNotFound):
		respondWithError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain_user.ErrUserAlreadyExists),
		errors.Is(err, domain_user.ErrUsernameAlreadyExists):
		respondWithError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain_user.ErrEmptyEmail),
		errors.Is(err, domain_user.ErrInvalidEmail),
		errors.Is(err, domain_user.ErrEmptyUsername),
		errors.Is(err, domain_user.ErrUsernameTooLong),
		errors.Is(err, domain_user.ErrWeakPassword):
		respondWithError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Errorf("unexpected error inside auth handler: %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")
	}
}
