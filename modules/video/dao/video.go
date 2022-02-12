package dao

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Video struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Width     uint32             `bson:"width,omitempty"`
	Height    uint32             `bson:"height,omitempty"`
	Size      uint64             `bson:"size,omitempty"`
	Duration  float64            `bson:"duration,omitempty"`
	URL       string             `bson:"url,omitempty"`
	Status    string             `bson:"status,omitempty"`
	Variants  map[string]string  `bson:"variants,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

type VideoDAO interface {
	Get(ctx context.Context, id primitive.ObjectID) (*Video, error)
	List(ctx context.Context, limit, skip int64) ([]*Video, error)
	Create(ctx context.Context, video *Video) (primitive.ObjectID, error)
	Update(ctx context.Context, video *Video) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var (
	ErrVideoNotFound = errors.New("video not found")
)
