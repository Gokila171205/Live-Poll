package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/repository"
	"live-poll-backend/internal/services"
)

type mockUserRepo struct {
	usersByEmail map[string]*models.User
	usersByID    map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*models.User),
		usersByID:    make(map[string]*models.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	normalized := strings.ToLower(user.Email)
	if _, exists := m.usersByEmail[normalized]; exists {
		return repository.ErrDuplicateEmail
	}
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	m.usersByEmail[normalized] = user
	m.usersByID[user.ID.Hex()] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	normalized := strings.ToLower(email)
	u, exists := m.usersByEmail[normalized]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	u, exists := m.usersByID[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) EnsureIndexes(ctx context.Context) error {
	return nil
}

func TestAuthService_Signup(t *testing.T) {
	repo := newMockUserRepo()
	authService := services.NewAuthService(repo, "test-secret-key", 24)

	ctx := context.Background()

	// 1. Success signup
	resp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "test@example.com",
		Password: "strongPassword123!",
	})
	if err != nil {
		t.Fatalf("unexpected signup error: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", resp.User.Email)
	}

	// Verify stored password was hashed with bcrypt
	savedUser, err := repo.FindByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("failed to find saved user: %v", err)
	}
	if savedUser.PasswordHash == "strongPassword123!" {
		t.Fatal("password must not be stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte("strongPassword123!")); err != nil {
		t.Errorf("stored password hash failed bcrypt verification: %v", err)
	}

	// 2. Duplicate email signup
	_, err = authService.Signup(ctx, models.SignupRequest{
		Email:    "test@example.com",
		Password: "anotherPassword123",
	})
	if !errors.Is(err, services.ErrEmailAlreadyInUse) {
		t.Errorf("expected ErrEmailAlreadyInUse for duplicate email, got: %v", err)
	}

	// 3. Short password validation (< 8 chars)
	_, err = authService.Signup(ctx, models.SignupRequest{
		Email:    "short@example.com",
		Password: "short",
	})
	if !errors.Is(err, services.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for short password, got: %v", err)
	}

	// 4. Invalid email format
	_, err = authService.Signup(ctx, models.SignupRequest{
		Email:    "not-an-email",
		Password: "strongPassword123!",
	})
	if !errors.Is(err, services.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for malformed email, got: %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockUserRepo()
	authService := services.NewAuthService(repo, "test-secret-key", 24)
	ctx := context.Background()

	_, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "loginuser@example.com",
		Password: "correctPassword123",
	})
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	// 1. Successful login
	resp, err := authService.Login(ctx, models.LoginRequest{
		Email:    "loginuser@example.com",
		Password: "correctPassword123",
	})
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected JWT token on successful login")
	}

	// 2. Wrong password
	_, err = authService.Login(ctx, models.LoginRequest{
		Email:    "loginuser@example.com",
		Password: "wrongPassword",
	})
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for wrong password, got: %v", err)
	}

	// 3. Nonexistent email
	_, err = authService.Login(ctx, models.LoginRequest{
		Email:    "unknown@example.com",
		Password: "somePassword123",
	})
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for unknown user, got: %v", err)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := newMockUserRepo()
	authService := services.NewAuthService(repo, "test-secret-key", 24)
	ctx := context.Background()

	resp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "tokenuser@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to signup: %v", err)
	}

	// Valid token
	claims, err := authService.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("expected token validation to succeed, got: %v", err)
	}
	if claims.UserID != resp.User.ID {
		t.Errorf("expected claims UserID to be '%s', got '%s'", resp.User.ID, claims.UserID)
	}

	// Tampered token
	_, err = authService.ValidateToken(resp.Token + "tampered")
	if !errors.Is(err, services.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for tampered token, got: %v", err)
	}

	// Service with different secret
	otherAuthService := services.NewAuthService(repo, "different-secret", 24)
	_, err = otherAuthService.ValidateToken(resp.Token)
	if !errors.Is(err, services.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for token signed with different key, got: %v", err)
	}
}

func TestAuthService_ExpiredToken(t *testing.T) {
	repo := newMockUserRepo()
	// Short 20ms expiration
	expiredService := services.NewAuthServiceWithDuration(repo, "test-secret", 20*time.Millisecond)
	ctx := context.Background()

	resp, err := expiredService.Signup(ctx, models.SignupRequest{
		Email:    "expired@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to signup: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	_, err = expiredService.ValidateToken(resp.Token)
	if !errors.Is(err, services.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for expired token, got: %v", err)
	}
}
