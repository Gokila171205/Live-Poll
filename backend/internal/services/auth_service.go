package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/repository"
)

var (
	// ErrInvalidCredentials indicates incorrect email or password during login.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrEmailAlreadyInUse indicates that an account already exists with the given email.
	ErrEmailAlreadyInUse = errors.New("email is already registered")
	// ErrInvalidInput indicates validation failure for input fields.
	ErrInvalidInput = errors.New("invalid input data")
	// ErrInvalidToken indicates the JWT is invalid or malformed.
	ErrInvalidToken = errors.New("invalid or expired token")
)

// AuthService defines business operations for authentication and authorization.
type AuthService interface {
	Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error)
	ValidateToken(tokenStr string) (*models.JWTClaims, error)
	GetUserByID(ctx context.Context, userID string) (*models.UserResponse, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtSecret  []byte
	jwtExpTime time.Duration
}

// NewAuthService creates a new AuthService instance with injected dependencies and config.
func NewAuthService(userRepo repository.UserRepository, jwtSecret string, expHours int) AuthService {
	if expHours <= 0 {
		expHours = 24
	}
	return NewAuthServiceWithDuration(userRepo, jwtSecret, time.Duration(expHours)*time.Hour)
}

// NewAuthServiceWithDuration creates a new AuthService with a specific token duration.
func NewAuthServiceWithDuration(userRepo repository.UserRepository, jwtSecret string, duration time.Duration) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		jwtExpTime: duration,
	}
}

// Signup registers a new user with hashed password and returns an auth token.
func (s *authService) Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	// Verify email is not already registered
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyInUse
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &models.User{
		Email:        email,
		PasswordHash: string(hashedBytes),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyInUse
		}
		return nil, fmt.Errorf("failed to persist user: %w", err)
	}

	token, err := s.generateToken(newUser.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	userResp := newUser.ToResponse()
	return &models.AuthResponse{
		Token: token,
		User:  userResp,
	}, nil
}

// Login verifies credentials and issues a signed JWT.
func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to look up user: %w", err)
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	userResp := user.ToResponse()
	return &models.AuthResponse{
		Token: token,
		User:  userResp,
	}, nil
}

// ValidateToken parses and verifies the authenticity of a JWT string.
func (s *authService) ValidateToken(tokenStr string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.UserID == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetUserByID looks up a user and returns a safe response representation.
func (s *authService) GetUserByID(ctx context.Context, userID string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) generateToken(userID string) (string, error) {
	now := time.Now().UTC()
	claims := models.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if len(email) > 254 {
		return errors.New("email cannot exceed 254 characters")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email address format")
	}
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return errors.New("email must contain a valid domain name")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(password) > 72 {
		return errors.New("password cannot exceed 72 characters (bcrypt limit)")
	}
	return nil
}
