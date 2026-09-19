package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"live-poll-backend/internal/middleware"
	"live-poll-backend/internal/models"
	"live-poll-backend/internal/services"
)

type mockAuthService struct {
	validateFn func(tokenStr string) (*models.JWTClaims, error)
}

func (m *mockAuthService) Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthService) ValidateToken(tokenStr string) (*models.JWTClaims, error) {
	if m.validateFn != nil {
		return m.validateFn(tokenStr)
	}
	return nil, services.ErrInvalidToken
}
func (m *mockAuthService) GetUserByID(ctx context.Context, userID string) (*models.UserResponse, error) {
	return nil, nil
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockAuthService{}
	router := gin.New()
	router.Use(middleware.AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Errorf("expected success to be false")
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", resp.Error.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockAuthService{}
	router := gin.New()
	router.Use(middleware.AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "TokenNotBearer")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedUserID := "user12345"
	mockSvc := &mockAuthService{
		validateFn: func(tokenStr string) (*models.JWTClaims, error) {
			if tokenStr == "valid-sample-token" {
				return &models.JWTClaims{
					UserID: expectedUserID,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: expectedUserID,
					},
				}, nil
			}
			return nil, services.ErrInvalidToken
		},
	}

	router := gin.New()
	router.Use(middleware.AuthMiddleware(mockSvc))
	router.GET("/protected", func(c *gin.Context) {
		uid, ok := middleware.GetAuthenticatedUserID(c)
		if !ok || uid != expectedUserID {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"userId": uid})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-sample-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
