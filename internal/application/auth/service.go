package auth

import (
	"Grippy/internal/domain/auth"
	"Grippy/internal/domain/user"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *auth.RefreshSession) error
	GetSessionByToken(ctx context.Context, token string) (*auth.RefreshSession, error)
	DeleteSession(ctx context.Context, token string) error
}

type UserRepository interface {
	Save(ctx context.Context, u *user.User) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
}

type Service struct {
	sessionRepo   SessionRepository
	userRepo      UserRepository
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type jwtClaims struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func NewService(
	sessionRepo SessionRepository,
	userRepo UserRepository,
	accessSecret string,
	refreshSecret string,
	accessTTL, refreshTTL time.Duration,
) *Service {
	return &Service{
		sessionRepo:   sessionRepo,
		userRepo:      userRepo,
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *Service) SignUp(ctx context.Context, input SignUpInput) (*user.User, error) {
	if len(input.Password) < 6 {
		return nil, user.ErrWeakPassword
	}

	if _, err := s.userRepo.GetByEmail(ctx, input.Email); err == nil {
		return nil, user.ErrUserAlreadyExists
	}

	if _, err := s.userRepo.GetByUsername(ctx, input.Username); err == nil {
		return nil, user.ErrUsernameAlreadyExists
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	newUser, err := user.NewUser(input.Email, input.Username, string(hashBytes))

	if err != nil {
		return nil, err
	}

	return s.userRepo.Save(ctx, newUser)
}
func (s *Service) SignIn(ctx context.Context, input SignInInput) (*auth.TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, input.Email)

	if err != nil {
		return nil, user.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash()), []byte(input.Password)); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	return s.GenerateTokenPair(ctx, u.ID(), u.Username())
}

func (s *Service) GenerateTokenPair(ctx context.Context, id uuid.UUID, username string) (*auth.TokenPair, error) {
	now := time.Now()

	accessClaims := jwtClaims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessTokenObj.SignedString(s.accessSecret)

	if err != nil {
		return nil, err
	}

	refreshClaims := jwtClaims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshTokenObj.SignedString(s.refreshSecret)

	if err != nil {
		return nil, err
	}

	session := auth.NewRefreshSession(id, refreshTokenStr, now, now.Add(s.refreshTTL))

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (*auth.TokenPair, error) {
	if input.RefreshToken == "" {
		return nil, auth.ErrInvalidToken
	}

	claims, err := s.ParseRefreshToken(input.RefreshToken)

	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	session, err := s.sessionRepo.GetSessionByToken(ctx, input.RefreshToken)

	if err != nil {
		return nil, auth.ErrSessionNotFound
	}

	if session.IsExpired() {
		_ = s.sessionRepo.DeleteSession(ctx, input.RefreshToken)
		return nil, auth.ErrInvalidToken
	}

	if err := s.sessionRepo.DeleteSession(ctx, input.RefreshToken); err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(ctx, claims.ID, claims.Username)
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken == "" {
		return auth.ErrInvalidToken
	}

	return s.sessionRepo.DeleteSession(ctx, input.RefreshToken)
}

func (s *Service) ParseAccessToken(tokenStr string) (*auth.Claims, error) {
	return s.parseToken(tokenStr, s.accessSecret)
}

func (s *Service) ParseRefreshToken(tokenStr string) (*auth.Claims, error) {
	return s.parseToken(tokenStr, s.refreshSecret)
}

func (s *Service) ValidateRefreshToken(ctx context.Context, tokenStr string) error {
	if _, err := s.ParseRefreshToken(tokenStr); err != nil {
		return auth.ErrInvalidToken
	}

	session, err := s.sessionRepo.GetSessionByToken(ctx, tokenStr)

	if err != nil {
		return auth.ErrSessionNotFound
	}

	if session.IsExpired() {
		_ = s.sessionRepo.DeleteSession(ctx, tokenStr)
		return auth.ErrInvalidToken
	}

	return nil
}

func (s *Service) parseToken(tokenStr string, secret []byte) (*auth.Claims, error) {
	var claims jwtClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	return &auth.Claims{
		ID:        claims.ID,
		Username:  claims.Username,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
