package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ContextKeyUserID is the key used to store the verified user ID in Gin context.
const ContextKeyUserID = "authenticatedUserID"

// User represents the persistent user entity in the database.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"password_hash" json:"-"`
	CreatedAt    time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updatedAt"`
}

// UserResponse defines the safe public representation of a user without sensitive data.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToResponse converts a User domain model into a safe UserResponse DTO.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID.Hex(),
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// SignupRequest defines the incoming payload for user registration.
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest defines the incoming payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

// AuthResponse defines the response payload returned on successful authentication.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// JWTClaims encapsulates custom claims embedded in signed JSON Web Tokens.
type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}
