package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"live-poll-backend/internal/database"
	"live-poll-backend/internal/models"
)

var (
	// ErrPollNotFound is returned when a poll cannot be located by ID.
	ErrPollNotFound = errors.New("poll not found")
	// ErrUnauthorizedAction is returned when a user attempts an action on a poll they do not own.
	ErrUnauthorizedAction = errors.New("unauthorized action on poll")
	// ErrAlreadyVoted is returned when a voter attempts to vote again on the same poll.
	ErrAlreadyVoted = errors.New("voter has already voted in this poll")
)

// PollRepository defines persistence operations for polls and persistent votes.
type PollRepository interface {
	Create(ctx context.Context, poll *models.Poll) error
	FindByID(ctx context.Context, id string) (*models.Poll, error)
	FindByCreatorID(ctx context.Context, creatorID string) ([]*models.Poll, error)
	Delete(ctx context.Context, id string, creatorID string) error
	SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error
	EnsureIndexes(ctx context.Context) error
	RecordVote(ctx context.Context, vote *models.Vote) error
	HasVoterVoted(ctx context.Context, pollID, voterID string) (bool, error)
}

type mongoPollRepository struct {
	mongo *database.MongoDB
}

// NewPollRepository initializes a MongoDB-backed PollRepository.
func NewPollRepository(mongo *database.MongoDB) PollRepository {
	return &mongoPollRepository{mongo: mongo}
}

func (r *mongoPollRepository) getPollsCollection() (*mongo.Collection, error) {
	if r.mongo == nil || r.mongo.Database == nil {
		return nil, database.ErrDatabaseNotInitialized
	}
	return r.mongo.Database.Collection("polls"), nil
}

func (r *mongoPollRepository) getVotesCollection() (*mongo.Collection, error) {
	if r.mongo == nil || r.mongo.Database == nil {
		return nil, database.ErrDatabaseNotInitialized
	}
	return r.mongo.Database.Collection("votes"), nil
}

// EnsureIndexes creates indexes on polls and votes collections.
func (r *mongoPollRepository) EnsureIndexes(ctx context.Context) error {
	pollsColl, err := r.getPollsCollection()
	if err != nil {
		return err
	}

	pollIndexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "creator_id", Value: 1}},
			Options: options.Index().SetName("idx_creator_id"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
	}

	_, err = pollsColl.Indexes().CreateMany(ctx, pollIndexModels)
	if err != nil {
		return fmt.Errorf("failed to create poll indexes: %w", err)
	}

	votesColl, err := r.getVotesCollection()
	if err != nil {
		return err
	}

	// Drop legacy index if present so partial filter expression can be safely created
	_ = votesColl.Indexes().DropOne(ctx, "unique_poll_user")

	voteIndexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_poll_voter"),
		},
		{
			Keys: bson.D{{Key: "poll_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "user_id", Value: bson.D{{Key: "$gt", Value: ""}}},
				}).
				SetName("unique_poll_user_partial"),
		},
	}

	_, err = votesColl.Indexes().CreateMany(ctx, voteIndexModels)
	if err != nil {
		return fmt.Errorf("failed to create vote indexes: %w", err)
	}

	return nil
}

// Create inserts a new poll into MongoDB.
func (r *mongoPollRepository) Create(ctx context.Context, poll *models.Poll) error {
	coll, err := r.getPollsCollection()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if poll.ID.IsZero() {
		poll.ID = bson.NewObjectID()
	}
	poll.CreatedAt = now
	poll.UpdatedAt = now
	poll.IsActive = true

	_, err = coll.InsertOne(ctx, poll)
	if err != nil {
		return fmt.Errorf("failed to insert poll: %w", err)
	}

	return nil
}

// FindByID retrieves a poll by its MongoDB ObjectID string.
func (r *mongoPollRepository) FindByID(ctx context.Context, id string) (*models.Poll, error) {
	coll, err := r.getPollsCollection()
	if err != nil {
		return nil, err
	}

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrPollNotFound
	}

	var poll models.Poll
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, fmt.Errorf("failed to find poll: %w", err)
	}

	return &poll, nil
}

// FindByCreatorID retrieves all polls created by a specific user.
func (r *mongoPollRepository) FindByCreatorID(ctx context.Context, creatorID string) ([]*models.Poll, error) {
	coll, err := r.getPollsCollection()
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := coll.Find(ctx, bson.M{"creator_id": creatorID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query polls by creator: %w", err)
	}
	defer cursor.Close(ctx)

	polls := make([]*models.Poll, 0)
	for cursor.Next(ctx) {
		var poll models.Poll
		if err := cursor.Decode(&poll); err != nil {
			return nil, fmt.Errorf("failed to decode poll: %w", err)
		}
		polls = append(polls, &poll)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while reading polls: %w", err)
	}

	return polls, nil
}

// Delete removes a poll if and only if the requesting user is the creator.
func (r *mongoPollRepository) Delete(ctx context.Context, id string, creatorID string) error {
	coll, err := r.getPollsCollection()
	if err != nil {
		return err
	}

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrPollNotFound
	}

	// Verify existence and ownership
	var poll models.Poll
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrPollNotFound
		}
		return fmt.Errorf("failed to fetch poll for deletion: %w", err)
	}

	if poll.CreatorID != creatorID {
		return ErrUnauthorizedAction
	}

	res, err := coll.DeleteOne(ctx, bson.M{"_id": objID, "creator_id": creatorID})
	if err != nil {
		return fmt.Errorf("failed to delete poll: %w", err)
	}

	if res.DeletedCount == 0 {
		return ErrPollNotFound
	}

	return nil
}

// SetPollStatus updates the active state of a poll (creator-only).
func (r *mongoPollRepository) SetPollStatus(ctx context.Context, id string, creatorID string, isActive bool) error {
	coll, err := r.getPollsCollection()
	if err != nil {
		return err
	}

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrPollNotFound
	}

	var poll models.Poll
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrPollNotFound
		}
		return fmt.Errorf("failed to fetch poll for status update: %w", err)
	}

	if poll.CreatorID != creatorID {
		return ErrUnauthorizedAction
	}

	filter := bson.M{"_id": objID, "creator_id": creatorID}
	update := bson.M{
		"$set": bson.M{
			"is_active":  isActive,
			"updated_at": time.Now().UTC(),
		},
	}

	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update poll status: %w", err)
	}

	return nil
}

// RecordVote atomically persists the vote document and increments the option count.
// Enforces race-condition safe uniqueness via the unique index on {poll_id, voter_id}.
func (r *mongoPollRepository) RecordVote(ctx context.Context, vote *models.Vote) error {
	votesColl, err := r.getVotesCollection()
	if err != nil {
		return err
	}
	pollsColl, err := r.getPollsCollection()
	if err != nil {
		return err
	}

	if vote.ID.IsZero() {
		vote.ID = bson.NewObjectID()
	}
	if vote.VotedAt.IsZero() {
		vote.VotedAt = time.Now().UTC()
	}

	// 1. Insert vote into votes collection (fails with duplicate key error if voter already voted)
	_, err = votesColl.InsertOne(ctx, vote)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrAlreadyVoted
		}
		return fmt.Errorf("failed to insert vote record: %w", err)
	}

	// 2. Increment persistent vote count on the matching option in the poll document
	objID, err := bson.ObjectIDFromHex(vote.PollID)
	if err != nil {
		return ErrPollNotFound
	}

	filter := bson.M{"_id": objID, "options.id": vote.OptionID}
	update := bson.M{
		"$inc": bson.M{"options.$.vote_count": 1},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}

	_, err = pollsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update poll persistent option count: %w", err)
	}

	return nil
}

// HasVoterVoted checks if the specified voter identifier has already cast a vote in the given poll.
func (r *mongoPollRepository) HasVoterVoted(ctx context.Context, pollID, voterID string) (bool, error) {
	if voterID == "" {
		return false, nil
	}

	votesColl, err := r.getVotesCollection()
	if err != nil {
		return false, err
	}

	count, err := votesColl.CountDocuments(ctx, bson.M{"poll_id": pollID, "voter_id": voterID})
	if err != nil {
		return false, fmt.Errorf("failed to check existing voter record: %w", err)
	}

	return count > 0, nil
}
