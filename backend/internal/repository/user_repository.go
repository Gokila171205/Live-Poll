package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"live-poll-backend/internal/database"
	"live-poll-backend/internal/models"
)

var (
	// ErrUserNotFound is returned when a requested user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateEmail is returned when an email is already registered.
	ErrDuplicateEmail = errors.New("email already exists")
)

// UserRepository defines database interactions for users.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	EnsureIndexes(ctx context.Context) error
}

type mongoUserRepository struct {
	mongo *database.MongoDB
}

// NewUserRepository creates an instance of UserRepository backed by MongoDB.
func NewUserRepository(mongo *database.MongoDB) UserRepository {
	return &mongoUserRepository{mongo: mongo}
}

func (r *mongoUserRepository) getCollection() (*mongo.Collection, error) {
	if r.mongo == nil || r.mongo.Database == nil {
		return nil, database.ErrDatabaseNotInitialized
	}
	return r.mongo.Database.Collection("users"), nil
}

// EnsureIndexes creates a unique index on the email field.
func (r *mongoUserRepository) EnsureIndexes(ctx context.Context) error {
	coll, err := r.getCollection()
	if err != nil {
		return err
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_user_email"),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create unique index on users.email: %w", err)
	}

	return nil
}

// Create inserts a new user record into the database.
func (r *mongoUserRepository) Create(ctx context.Context, user *models.User) error {
	coll, err := r.getCollection()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err = coll.InsertOne(ctx, user)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// FindByEmail searches for a user matching the provided email address.
func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	coll, err := r.getCollection()
	if err != nil {
		return nil, err
	}

	var user models.User
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	err = coll.FindOne(ctx, bson.M{"email": normalizedEmail}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return &user, nil
}

// FindByID retrieves a user record by its hex string ObjectID.
func (r *mongoUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	coll, err := r.getCollection()
	if err != nil {
		return nil, err
	}

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by ID: %w", err)
	}

	return &user, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// Check mongo.IsDuplicateKeyError or write exception code 11000
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, writeError := range writeErr.WriteErrors {
			if writeError.Code == 11000 {
				return true
			}
		}
	}
	var commandErr mongo.CommandError
	if errors.As(err, &commandErr) && commandErr.Code == 11000 {
		return true
	}
	return strings.Contains(err.Error(), "E11000 duplicate key error")
}
