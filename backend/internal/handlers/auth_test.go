package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"live-poll-backend/internal/handlers"
	"live-poll-backend/internal/middleware"
	"live-poll-backend/internal/models"
	"live-poll-backend/internal/repository"
	"live-poll-backend/internal/services"
)

type memoryUserRepo struct {
	usersByEmail map[string]*models.User
	usersByID    map[string]*models.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{
		usersByEmail: make(map[string]*models.User),
		usersByID:    make(map[string]*models.User),
	}
}

func (m *memoryUserRepo) Create(ctx context.Context, user *models.User) error {
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

func (m *memoryUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u, exists := m.usersByEmail[strings.ToLower(email)]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *memoryUserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	u, exists := m.usersByID[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *memoryUserRepo) EnsureIndexes(ctx context.Context) error {
	return nil
}

func setupTestRouter() (*gin.Engine, services.AuthService) {
	gin.SetMode(gin.TestMode)

	userRepo := newMemoryUserRepo()
	authService := services.NewAuthService(userRepo, "unit-test-jwt-secret", 24)
	authHandler := handlers.NewAuthHandler(authService)

	pollRepo := newMemoryPollRepo()
	pollService := services.NewPollService(pollRepo, nil, nil)
	pollHandler := handlers.NewPollHandler(pollService)

	router := gin.New()
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.GET("/my/polls", pollHandler.GetMyPolls)
		}
	}

	return router, authService
}

func TestAuthEndpoints_FullFlow(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Signup - Valid user
	signupPayload := `{"email": "alice@example.com", "password": "securePassword123"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(signupPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var signupResp struct {
		Success bool                `json:"success"`
		Data    models.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &signupResp); err != nil {
		t.Fatalf("failed to decode signup response: %v", err)
	}

	if !signupResp.Success || signupResp.Data.Token == "" {
		t.Fatalf("signup response missing token or success flag: %+v", signupResp)
	}

	token := signupResp.Data.Token
	userID := signupResp.Data.User.ID

	// Verify no password hash is exposed in the response
	if strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), "passwordHash") {
		t.Errorf("response exposed password hash")
	}

	// 2. Signup - Duplicate email
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(signupPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate email, got %d", w.Code)
	}

	// 3. Signup - Invalid input (short password)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(`{"email":"bob@example.com","password":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for short password, got %d", w.Code)
	}

	// 4. Login - Valid credentials
	loginPayload := `{"email": "alice@example.com", "password": "securePassword123"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid login, got %d: %s", w.Code, w.Body.String())
	}

	var loginResp struct {
		Success bool                `json:"success"`
		Data    models.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if loginResp.Data.Token == "" {
		t.Fatal("expected token in login response")
	}

	// 5. Login - Wrong password
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"alice@example.com","password":"wrongPassword"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for wrong password, got %d", w.Code)
	}

	// 6. Access protected endpoint WITHOUT token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without token, got %d", w.Code)
	}

	// 7. Access protected endpoint WITH valid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for authenticated /me, got %d: %s", w.Code, w.Body.String())
	}

	var meResp struct {
		Success bool                `json:"success"`
		Data    models.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("failed to decode /me response: %v", err)
	}
	if meResp.Data.ID != userID || meResp.Data.Email != "alice@example.com" {
		t.Errorf("mismatched user profile from token: %+v", meResp.Data)
	}

	// 8. Access protected poll endpoint WITHOUT token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/my/polls", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for /api/my/polls without token, got %d", w.Code)
	}

	// 9. Access protected poll endpoint WITH valid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/my/polls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/my/polls with token, got %d: %s", w.Code, w.Body.String())
	}

	var pollsResp struct {
		Success bool                  `json:"success"`
		Data    []models.PollResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &pollsResp); err != nil {
		t.Fatalf("failed to decode polls response: %v", err)
	}
	if len(pollsResp.Data) != 0 {
		t.Errorf("expected empty polls list, got %d", len(pollsResp.Data))
	}
}
