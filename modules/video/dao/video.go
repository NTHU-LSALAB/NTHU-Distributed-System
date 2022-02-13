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
	Create(ctx context.Context, video *Video) error
	Update(ctx context.Context, video *Video) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

var (
	ErrVideoNotFound = errors.New("video not found")
)

// newFakeVideo returns a fake video instance with random
// id that is useful for testing
func newFakeVideo() *Video {
	id := primitive.NewObjectID()

	// Note that timestamp is hard to test equally,
	// so ignore the `createdAt` and `updatedAt` field

	return &Video{
		ID:       id,
		Width:    800,
		Height:   600,
		Size:     144000,
		Duration: 10.234,
		URL:      "https://storage.example.com/videos/" + id.Hex() + ".mp4",
		Status:   "SUCCESS",
		Variants: map[string]string{
			"1080p": "https://storage.example.com/videos/" + id.Hex() + "-1080p.mp4",
			"720p":  "https://storage.example.com/videos/" + id.Hex() + "-720p.mp4",
		},
	}
}
