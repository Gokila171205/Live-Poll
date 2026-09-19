package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PollOption represents a single choice in a poll.
type PollOption struct {
	ID        string `bson:"id" json:"id"`
	Text      string `bson:"text" json:"text"`
	VoteCount int64  `bson:"vote_count" json:"voteCount"`
}

// Poll represents a poll record stored in the database.
type Poll struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatorID string        `bson:"creator_id" json:"creatorId"`
	Question  string        `bson:"question" json:"question"`
	Options   []PollOption  `bson:"options" json:"options"`
	IsActive  bool          `bson:"is_active" json:"isActive"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updatedAt"`
}

// CreatePollRequest represents the payload for creating a new poll.
type CreatePollRequest struct {
	Question string   `json:"question" binding:"required,min=3,max=300"`
	Options  []string `json:"options" binding:"required,min=2,max=10,dive,min=1,max=100"`
}

// PollResponse represents the sanitized JSON output returned by poll APIs.
type PollResponse struct {
	ID        string       `json:"id"`
	CreatorID string       `json:"creatorId"`
	Question  string       `json:"question"`
	Options   []PollOption `json:"options"`
	IsActive  bool         `json:"isActive"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// ToResponse formats a Poll database document into a clean API response.
func (p *Poll) ToResponse() PollResponse {
	return PollResponse{
		ID:        p.ID.Hex(),
		CreatorID: p.CreatorID,
		Question:  p.Question,
		Options:   p.Options,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// CastVoteRequest defines the payload for voting on a poll option.
type CastVoteRequest struct {
	OptionID string `json:"optionId" binding:"required,min=1,max=64"`
}

// SetPollStatusRequest defines the payload for activating or closing a poll.
type SetPollStatusRequest struct {
	IsActive bool `json:"isActive"`
}

// VoteResultOption provides the vote tally and calculated percentage for an option.
type VoteResultOption struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	VoteCount  int64   `json:"voteCount"`
	Percentage float64 `json:"percentage"`
}

// PollResultsResponse encapsulates the live results response for a poll.
type PollResultsResponse struct {
	PollID     string             `json:"pollId"`
	Question   string             `json:"question"`
	IsActive   bool               `json:"isActive"`
	TotalVotes int64              `json:"totalVotes"`
	Options    []VoteResultOption `json:"options"`
}

// PollOptionResult represents an individual option's count in a live event broadcast.
type PollOptionResult struct {
	OptionID   string  `json:"optionId"`
	Text       string  `json:"text"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// PollUpdatedEvent defines the standardized WebSocket payload broadcast when a poll is updated.
type PollUpdatedEvent struct {
	Type       string             `json:"type"`
	PollID     string             `json:"pollId"`
	Question   string             `json:"question"`
	IsActive   bool               `json:"isActive"`
	Results    []PollOptionResult `json:"results"`
	TotalVotes int64              `json:"totalVotes"`
}

// ToEvent converts a PollResultsResponse into a standardized PollUpdatedEvent.
func (r *PollResultsResponse) ToEvent() PollUpdatedEvent {
	results := make([]PollOptionResult, len(r.Options))
	for i, o := range r.Options {
		results[i] = PollOptionResult{
			OptionID:   o.ID,
			Text:       o.Text,
			Count:      o.VoteCount,
			Percentage: o.Percentage,
		}
	}
	return PollUpdatedEvent{
		Type:       "poll_results_updated",
		PollID:     r.PollID,
		Question:   r.Question,
		IsActive:   r.IsActive,
		Results:    results,
		TotalVotes: r.TotalVotes,
	}
}

// Vote represents a persistent record of an individual vote cast in MongoDB.
type Vote struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID   string        `bson:"poll_id" json:"pollId"`
	OptionID string        `bson:"option_id" json:"optionId"`
	VoterID  string        `bson:"voter_id" json:"voterId"`
	UserID   string        `bson:"user_id,omitempty" json:"userId,omitempty"`
	VotedAt  time.Time     `bson:"voted_at" json:"votedAt"`
}
